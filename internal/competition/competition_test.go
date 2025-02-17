package competition

import (
	"testing"

	"github.com/SntrKslnn/matchmaking-service/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestAddingPlayers tests the adding of players to the competition
func TestAddingPlayers(t *testing.T) {
	competition := NewCompetition(
		1,
		CompetitionConfig{
			MaxPlayerCount: 10,
			MinPlayerCount: 2,
		},
		CompetitionLevelRange{
			Min: 1,
			Max: 10,
		},
	)

	competition.AddPlayer(model.PlayerData{ID: "test", Level: 1})
	competition.AddPlayer(model.PlayerData{ID: "test2", Level: 2})
	competition.AddPlayer(model.PlayerData{ID: "test3", Level: 3})
	competition.AddPlayer(model.PlayerData{ID: "test4", Level: 4})
	competition.AddPlayer(model.PlayerData{ID: "test5", Level: 5})
	competition.AddPlayer(model.PlayerData{ID: "test6", Level: 6})
	competition.AddPlayer(model.PlayerData{ID: "test7", Level: 7})
	competition.AddPlayer(model.PlayerData{ID: "test8", Level: 8})
	competition.AddPlayer(model.PlayerData{ID: "test9", Level: 9})

	assert.Equal(t, 9, competition.GetNumberOfJoinedPlayers())

}
