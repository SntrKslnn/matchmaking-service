package matchmaking

import (
	"time"

	"github.com/SntrKslnn/matchmaking-service/internal/competition"
	"github.com/SntrKslnn/matchmaking-service/internal/model"
)

type MatchmakingService interface {
	// HandlePlayerJoin handles a player's request to join matchmaking and returns a notification channel
	// that will receive updates about competition matching
	HandlePlayerJoin(playerData model.PlayerData) <-chan MatchMakingNotification
}

// MatchmakingConfig is the configuration for the matchmaking service
type MatchmakingConfig struct {
	LevelMatchingTolerance int
	MatchmakingTimeout     time.Duration
	CompetitionConfig      competition.CompetitionConfig
}

// MatchMakingNotification is a notification that is sent to the player to keep them updated about the matchmaking process
type MatchMakingNotification struct {
	CompetitionID int
	State         MatchmakingState
}

// State of the competition in matchmaking
type MatchmakingState string

const (
	// Indicates that the competition is open and accepting new players
	State_WaitingForPlayers MatchmakingState = "waiting_for_players"

	// Indicates that the competition has started
	State_Started MatchmakingState = "started"

	// Indicates that the competition has been aborted
	State_Aborted MatchmakingState = "aborted"
)

// NewMatchmakingService creates a new matchmaking service
// @param config the configuration of the matchmaking service
// @return a new matchmaking service
func NewMatchmakingService(config MatchmakingConfig) MatchmakingService {
	return newMatchmakingService(config)
}

func (m *matchmakingService) HandlePlayerJoin(playerData model.PlayerData) <-chan MatchMakingNotification {
	return m.handlePlayerJoin(playerData)
}
