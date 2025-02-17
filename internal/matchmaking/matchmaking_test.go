package matchmaking

import (
	"testing"
	"time"

	"github.com/SntrKslnn/matchmaking-service/internal/competition"
	"github.com/SntrKslnn/matchmaking-service/internal/model"
	"github.com/stretchr/testify/assert"
)

type TestPlayer struct {
	model.PlayerData
	personalNotificationChannel <-chan MatchMakingNotification
}

func createTesUsers(playerData []model.PlayerData) []TestPlayer {
	testUsers := make([]TestPlayer, len(playerData))
	for i, player := range playerData {
		testUsers[i] = TestPlayer{PlayerData: player, personalNotificationChannel: nil}
	}
	return testUsers
}

func listenForCompetitionStartNotification(testPlayer TestPlayer, expectedNotification map[string]bool) {
	for notification := range testPlayer.personalNotificationChannel {
		if notification.State == State_Started {
			expectedNotification[testPlayer.ID] = true
		}
	}
}

func listenPlayerNotifications(testUsers []TestPlayer, expectedNotification map[string]bool) {
	for _, testPlayer := range testUsers {
		go listenForCompetitionStartNotification(testPlayer, expectedNotification)
	}
}

func TestMatchmakingService_HandlePlayerJoin(t *testing.T) {
	matchmakingService := newMatchmakingService(MatchmakingConfig{
		CompetitionConfig: competition.CompetitionConfig{
			MaxPlayerCount: 10,
			MinPlayerCount: 2,
		},
		MatchmakingTimeout:     3 * time.Second,
		LevelMatchingTolerance: 3,
	})

	testUsers := createTesUsers([]model.PlayerData{
		{ID: "test_user_1", Level: 1},
		{ID: "test_user_2", Level: 2},
	})

	playerReceivedExpectedNotification := map[string]bool{
		"test_user_1": false,
		"test_user_2": false,
	}

	go func() {
		joinPlayersToMatchmaking(matchmakingService, testUsers)
		listenPlayerNotifications(testUsers, playerReceivedExpectedNotification)
	}()

	assert.Eventually(t, func() bool {
		for _, expectedNotification := range playerReceivedExpectedNotification {
			if !expectedNotification {
				return false
			}
		}
		return true
	}, time.Second*100, time.Millisecond*100)
}

func TestMatchmakingService_OverlappingLevels(t *testing.T) {
	matchmakingService := newMatchmakingService(MatchmakingConfig{
		CompetitionConfig: competition.CompetitionConfig{
			MaxPlayerCount: 10,
			MinPlayerCount: 2,
		},
		MatchmakingTimeout:     3 * time.Second,
		LevelMatchingTolerance: 3,
	})

	minLevel, maxLevel := matchmakingService.getLevelRangeMatchmakingConfiguratedOverlap(model.PlayerData{ID: "test_user_1", Level: 1})
	assert.Equal(t, 1, minLevel)
	assert.Equal(t, 4, maxLevel)

	minLevel, maxLevel = matchmakingService.getLevelRangeMatchmakingConfiguratedOverlap(model.PlayerData{ID: "test_user_50", Level: 50})
	assert.Equal(t, 47, minLevel)
	assert.Equal(t, 53, maxLevel)

}

func TestMatchmakingService_NewCompetitionAreCreatedRespectingLevelDistributionOverlap(t *testing.T) {
	matchmakingService := newMatchmakingService(MatchmakingConfig{
		CompetitionConfig: competition.CompetitionConfig{
			MaxPlayerCount: 10,
			MinPlayerCount: 2,
		},
		MatchmakingTimeout:     3 * time.Second,
		LevelMatchingTolerance: 3,
	})

	go func() {
		testPlayers := []TestPlayer{
			{PlayerData: model.PlayerData{ID: "test_user_1", Level: 1}},
			{PlayerData: model.PlayerData{ID: "test_user_2", Level: 2}},
			{PlayerData: model.PlayerData{ID: "test_user_3", Level: 3}},

			{PlayerData: model.PlayerData{ID: "test_user_5", Level: 25}},
			{PlayerData: model.PlayerData{ID: "test_user_5", Level: 25}},
			{PlayerData: model.PlayerData{ID: "test_user_5", Level: 25}},

			{PlayerData: model.PlayerData{ID: "test_user_6", Level: 60}},
			{PlayerData: model.PlayerData{ID: "test_user_7", Level: 61}},
			{PlayerData: model.PlayerData{ID: "test_user_8", Level: 62}},
			{PlayerData: model.PlayerData{ID: "test_user_9", Level: 63}},
		}
		joinPlayersToMatchmaking(matchmakingService, testPlayers)
		listenPlayerNotifications(testPlayers, nil)
	}()

	assert.Eventually(t, func() bool {
		return len(matchmakingService.competitionsInMatchmaking) == 3
	}, 10*time.Second, 50*time.Millisecond)
}

func joinPlayersToMatchmaking(matchmakingService *matchmakingService, players []TestPlayer) {
	for i := range players {
		notificationChannel := matchmakingService.HandlePlayerJoin(players[i].PlayerData)
		players[i].personalNotificationChannel = notificationChannel
	}
}
