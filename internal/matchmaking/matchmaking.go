package matchmaking

import (
	"log/slog"
	"time"

	"github.com/SntrKslnn/matchmaking-service/internal/competition"
	"github.com/SntrKslnn/matchmaking-service/internal/model"
)

type playerInMatchmaking struct {
	model.PlayerData
	matchMakingNotificationChan chan MatchMakingNotification
}

type competitionData struct {
	competition.Competition
	timeoutCancel chan struct{}
}

type matchmakingService struct {
	nextCompetitionID int
	config            MatchmakingConfig

	playersInMatchmaking      map[string]playerInMatchmaking
	competitionsInMatchmaking map[int]competitionData

	competitionJoinRequests chan model.PlayerData

	stateMutationChan chan stateChangeNotification
}

type matchmakingStateChangeOrigin string

const (
	matchmakingStateChangeOrigin_PlayerAdd matchmakingStateChangeOrigin = "player_added"
	matchmakingStateChangeOrigin_Timeout   matchmakingStateChangeOrigin = "matchmaking_timeout"
)

type stateChangeNotification struct {
	origin      matchmakingStateChangeOrigin
	competition competition.Competition
	playerData  model.PlayerData
}

func newMatchmakingService(config MatchmakingConfig) *matchmakingService {
	matchmakingService := &matchmakingService{
		competitionsInMatchmaking: make(map[int]competitionData),
		playersInMatchmaking:      make(map[string]playerInMatchmaking),
		competitionJoinRequests:   make(chan model.PlayerData),
		nextCompetitionID:         1,
		config:                    config,
		stateMutationChan:         make(chan stateChangeNotification),
	}
	matchmakingService.start()
	return matchmakingService
}

// TODO This would need to be fixed. Noticed it very late that this is actually a race condition
// because MatchmakingService.HandlePlayerJoin can be called from different goroutines.
// It leads to situation where different goroutines can access concurrently m.playersInMatchmaking
// Better approach would be to mutate m.playersInMatchmaking from single goroutine always
// Other option would be to use mutex
func (m *matchmakingService) handlePlayerJoin(playerData model.PlayerData) <-chan MatchMakingNotification {
	if _, exists := m.playersInMatchmaking[playerData.ID]; !exists {
		matchMakingNotificationChan := make(chan MatchMakingNotification)
		m.playersInMatchmaking[playerData.ID] = playerInMatchmaking{
			PlayerData:                  playerData,
			matchMakingNotificationChan: matchMakingNotificationChan,
		}
	}
	m.competitionJoinRequests <- playerData
	return m.playersInMatchmaking[playerData.ID].matchMakingNotificationChan
}

func (m *matchmakingService) listenCompetitionJoinRequest() {
	for playerData := range m.competitionJoinRequests {
		m.updateMatchmakingWithPlayerJoinRequest(playerData)
	}
}

func (m *matchmakingService) getMatchMakingState(notificationOrigin matchmakingStateChangeOrigin, competition competition.Competition) MatchmakingState {
	maxPlayerCountReached := competition.GetNumberOfJoinedPlayers() == m.config.CompetitionConfig.MaxPlayerCount
	minPlayerCountReached := competition.GetNumberOfJoinedPlayers() >= m.config.CompetitionConfig.MinPlayerCount

	if notificationOrigin == matchmakingStateChangeOrigin_PlayerAdd {
		if maxPlayerCountReached {
			slog.Info("Max player count reached. Starting competition", "id", competition.GetID())
			// Competition is full, start it
			return State_Started
		}
		// Competition is not full, wait for players
		return State_WaitingForPlayers
	}

	// Competition has timeouted
	if minPlayerCountReached {
		slog.Info("Matchmaking timeouted. Min player count reached. Starting competition", "id", competition.GetID())
		// Competition has reached the minimum player count, start it
		return State_Started
	}

	// Competition has not reached the minimum player count, abort it
	slog.Info("Matchmaking timeouted. Min player count not reached. Aborting competition", "id", competition.GetID())
	return State_Aborted
}

func (m *matchmakingService) processMatchmakingStateMutation(stateChangeNotification stateChangeNotification) {

	competition := stateChangeNotification.competition
	notificationOrigin := stateChangeNotification.origin

	if notificationOrigin == matchmakingStateChangeOrigin_PlayerAdd {
		playerData := stateChangeNotification.playerData
		competition = m.handleAddingPlayerToCompetition(playerData)
	}

	competitionState := m.getMatchMakingState(notificationOrigin, competition)

	switch competitionState {
	case State_Started:
		m.startCompetition(competition)
	case State_Aborted:
		m.abortCompetition(competition)
	}
}

func (m *matchmakingService) listenCompetitionStatusCheckChan() {
	for stateChangeNotification := range m.stateMutationChan {
		m.processMatchmakingStateMutation(stateChangeNotification)
	}
}

// This is a naive implementation that iterates over all the competitions to find the first one that matches the player's level
// This is not efficient for a large number of competitions
// This could be improved by using more efficient data structures for constant time lookup
func (m *matchmakingService) findFirstCompetitionThatMatchesPlayerLevel(playerData model.PlayerData) (competition.Competition, bool) {
	for _, competitionData := range m.competitionsInMatchmaking {
		competition := competitionData.Competition
		if competition.IsPlayerLevelMatching(playerData) {
			return competition, true
		}
	}
	return nil, false
}

func (m *matchmakingService) findCompetitionForPlayer(playerData model.PlayerData) (competition.Competition, bool) {
	return m.findFirstCompetitionThatMatchesPlayerLevel(playerData)
}

func (m *matchmakingService) sendNotificationToPlayer(playerID string, matchMakingNotification MatchMakingNotification) {
	m.playersInMatchmaking[playerID].matchMakingNotificationChan <- matchMakingNotification
}

func (m *matchmakingService) addPlayerToCompetition(playerData model.PlayerData, competitionToAddPlayerTo competition.Competition) {
	competitionToAddPlayerTo.AddPlayer(playerData)

	go m.sendNotificationToPlayer(playerData.ID, MatchMakingNotification{
		CompetitionID: competitionToAddPlayerTo.GetID(),
		State:         State_WaitingForPlayers,
	})

	slog.Info(
		"Player joined competition",
		"id", competitionToAddPlayerTo.GetID(),
		"player_id", playerData.ID,
	)
}

// handleAddingPlayerToCompetition handles the adding of a player to a competition
// It will find a competition for the player or create a new one if no competition is found
// It will then add the player to the competition
// @param playerData the player to add to the competition
// @return the competition that the player was added to
func (m *matchmakingService) handleAddingPlayerToCompetition(playerData model.PlayerData) competition.Competition {
	competition, found := m.findCompetitionForPlayer(playerData)
	if !found {
		competition = m.createNewCompetition(playerData)
	}
	m.addPlayerToCompetition(playerData, competition)
	return competition
}

func (m *matchmakingService) sendStateMutationCommands(stateMutationCommand stateChangeNotification) {
	m.stateMutationChan <- stateMutationCommand
}

func (m *matchmakingService) updateMatchmakingWithPlayerJoinRequest(playerData model.PlayerData) {
	m.sendStateMutationCommands(stateChangeNotification{
		origin:     matchmakingStateChangeOrigin_PlayerAdd,
		playerData: playerData,
	})
}

func (m *matchmakingService) startTimeoutTimerForCompetition(competition competition.Competition) {

	select {
	case <-time.After(m.config.MatchmakingTimeout):
		slog.Info("Matchmaking timeouted. Checking for minimum player count", "id", competition.GetID())
		m.sendStateMutationCommands(stateChangeNotification{
			origin:      matchmakingStateChangeOrigin_Timeout,
			competition: competition,
		})
	case <-m.competitionsInMatchmaking[competition.GetID()].timeoutCancel:
		return
	}
}

func (m *matchmakingService) start() {
	go m.listenCompetitionStatusCheckChan()
	go m.listenCompetitionJoinRequest()
}

func (m *matchmakingService) getLevelRangeMatchmakingConfiguratedOverlap(playerData model.PlayerData) (int, int) {
	playerMinLevel := playerData.Level - m.config.LevelMatchingTolerance
	if playerMinLevel < 1 {
		playerMinLevel = 1
	}
	playerMaxLevel := playerData.Level + m.config.LevelMatchingTolerance
	return playerMinLevel, playerMaxLevel
}

func (m *matchmakingService) createNewCompetition(playerData model.PlayerData) competition.Competition {
	playerMinLevel, playerMaxLevel := m.getLevelRangeMatchmakingConfiguratedOverlap(playerData)

	competition := competition.NewCompetition(m.nextCompetitionID, m.config.CompetitionConfig, competition.CompetitionLevelRange{
		Min: playerMinLevel,
		Max: playerMaxLevel,
	})

	slog.Info("Creating new competition", "id", competition.GetID(), "min_level", playerMinLevel, "max_level", playerMaxLevel)

	m.competitionsInMatchmaking[competition.GetID()] = competitionData{
		Competition:   competition,
		timeoutCancel: make(chan struct{}),
	}

	go m.startTimeoutTimerForCompetition(competition)

	m.nextCompetitionID++

	return competition
}

func (m *matchmakingService) unregisterCompetitionFromMatchmakingStage(competition competition.Competition) {
	delete(m.competitionsInMatchmaking, competition.GetID())
	slog.Info("Deleted competition from pending competitions", "id", competition.GetID())
}

func (m *matchmakingService) unregisterPlayersFromMatchmakingStage(competition competition.Competition) {
	for _, player := range competition.GetPlayers() {
		delete(m.playersInMatchmaking, player.ID)
		slog.Info("Deleted player from matchmaking", "id", player.ID)
	}
}

func (m *matchmakingService) startCompetition(competition competition.Competition) {
	competition.Start()
	m.closeTimeoutCancelChannelForCompetition(competition)
	m.notifyPlayers(competition, State_Started)
	m.unregisterCompetitionFromMatchmakingStage(competition)
	m.unregisterPlayersFromMatchmakingStage(competition)
}

func (m *matchmakingService) abortCompetition(competition competition.Competition) {
	m.closeTimeoutCancelChannelForCompetition(competition)
	m.notifyPlayers(competition, State_Aborted)
	m.unregisterCompetitionFromMatchmakingStage(competition)
	m.unregisterPlayersFromMatchmakingStage(competition)
}

func (m *matchmakingService) closeTimeoutCancelChannelForCompetition(competition competition.Competition) {
	close(m.competitionsInMatchmaking[competition.GetID()].timeoutCancel)
}

func (m *matchmakingService) notifyPlayers(startedCompetition competition.Competition, state MatchmakingState) {
	for _, player := range startedCompetition.GetPlayers() {
		m.sendNotificationToPlayer(player.ID, MatchMakingNotification{
			CompetitionID: startedCompetition.GetID(),
			State:         state,
		})
	}
}
