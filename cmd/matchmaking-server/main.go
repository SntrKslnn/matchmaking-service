package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/SntrKslnn/matchmaking-service/internal/competition"
	"github.com/SntrKslnn/matchmaking-service/internal/matchmaking"
	"github.com/SntrKslnn/matchmaking-service/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "TCP server port")
	maxPlayers := flag.Int("max-players", 10, "Maximum number of players per competition")
	minPlayers := flag.Int("min-players", 2, "Minimum number of players to start competition")
	levelOverlap := flag.Int("level-matching-tolerance", 3, "Level overlap for matchmaking")
	timeout := flag.Duration("timeout", 20*time.Second, "Matchmaking timeout duration")
	flag.Parse()

	fmt.Printf("Starting TCP server on port %d with max players %d, min players %d, level overlap %d, and timeout %s\n", *port, *maxPlayers, *minPlayers, *levelOverlap, *timeout)

	matchMakingTcpServer := server.NewTCPServer(*port, matchmaking.MatchmakingConfig{
		CompetitionConfig: competition.CompetitionConfig{
			MaxPlayerCount: *maxPlayers,
			MinPlayerCount: *minPlayers,
		},
		MatchmakingTimeout:     *timeout,
		LevelMatchingTolerance: *levelOverlap,
	})

	matchMakingTcpServer.Start()

}
