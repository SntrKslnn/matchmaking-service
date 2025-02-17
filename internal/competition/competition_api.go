package competition

import (
	"github.com/SntrKslnn/matchmaking-service/internal/model"
)

type Competition interface {
	// AddPlayer adds a player to the competition
	// @param playerData the data of the player to add
	AddPlayer(playerData model.PlayerData)

	// IsPlayerLevelMatching checks if a player's level is within the competition's level range
	// @param playerData the data of the player to check
	// @return true if the player's level is within the competition's level range, false otherwise
	IsPlayerLevelMatching(playerData model.PlayerData) bool

	// GetID returns the id of the competition
	// @return the id of the competition
	GetID() int

	// GetPlayers returns the players of the competition
	// @return the players of the competition
	GetPlayers() map[string]model.PlayerData

	// GetNumberOfJoinedPlayers returns the number of joined players in the competition
	// @return the number of joined players in the competition
	GetNumberOfJoinedPlayers() int

	// Start starts the competition
	Start()
}

// CompetitionLevelRange represents the range of levels a player can be in to join a competition
type CompetitionLevelRange struct {
	Min int
	Max int
}

// CompetitionConfig represents the configuration of a competition
type CompetitionConfig struct {
	MaxPlayerCount int
	MinPlayerCount int
}

// CompetitionStateChangeNotification is a notification about a competition state change
type CompetitionStateChangeNotification bool

// Creates new competition
func NewCompetition(id int, config CompetitionConfig, playerLevelRange CompetitionLevelRange) Competition {
	return newCompetition(id, config, playerLevelRange)
}

func (c *competition) AddPlayer(playerData model.PlayerData) {
	c.addPlayer(playerData)
}

func (c *competition) IsPlayerLevelMatching(playerData model.PlayerData) bool {
	return c.isPlayerLevelMatching(playerData)
}

func (c *competition) GetID() int {
	return c.getID()
}

func (c *competition) GetNumberOfJoinedPlayers() int {
	return c.getNumberOfJoinedPlayers()
}

func (c *competition) GetPlayers() map[string]model.PlayerData {
	return c.getPlayers()
}

func (c *competition) Start() {
	c.start()
}
