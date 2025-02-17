package competition

import (
	"log/slog"

	"github.com/SntrKslnn/matchmaking-service/internal/model"
)

type competition struct {
	id               int
	config           CompetitionConfig
	playerLevelRange CompetitionLevelRange

	players map[string]model.PlayerData
}

func newCompetition(id int, config CompetitionConfig, playerLevelRange CompetitionLevelRange) *competition {
	return &competition{
		id:               id,
		config:           config,
		playerLevelRange: playerLevelRange,
		players:          make(map[string]model.PlayerData),
	}
}

func (c *competition) addPlayer(playerData model.PlayerData) {
	c.addPlayerToCompetition(playerData)
}

func (c *competition) isPlayerLevelMatching(playerData model.PlayerData) bool {
	return playerData.Level >= c.playerLevelRange.Min && playerData.Level <= c.playerLevelRange.Max
}

func (c *competition) addPlayerToCompetition(playerData model.PlayerData) {
	c.players[playerData.ID] = playerData
}

func (c *competition) getID() int {
	return c.id
}

func (c *competition) getPlayers() map[string]model.PlayerData {
	return c.players
}

func (c *competition) getNumberOfJoinedPlayers() int {
	return len(c.players)
}

func (c *competition) start() {
	slog.Info("Competition started", "id", c.id, "players", c.getPlayers())
}
