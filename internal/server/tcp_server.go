package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	"github.com/SntrKslnn/matchmaking-service/internal/matchmaking"
	"github.com/SntrKslnn/matchmaking-service/internal/model"
)

type MatchmakingTcpServer interface {
	// Start starts the TCP server
	Start() error

	// Stop stops the TCP server
	Stop() error
}

type tcpServer struct {
	listener           net.Listener
	port               int
	matchmakingService matchmaking.MatchmakingService
}

func NewTCPServer(port int, matchmakingConfig matchmaking.MatchmakingConfig) MatchmakingTcpServer {
	return &tcpServer{
		port:               port,
		matchmakingService: matchmaking.NewMatchmakingService(matchmakingConfig),
	}
}

func (s *tcpServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}
	s.listener = listener

	slog.Info("TCP Server listening.", "port", s.port)
	s.listenForConnections()
	return nil
}

func (s *tcpServer) listenForConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "error", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *tcpServer) handlePlayerJoinRequest(reader *bufio.Reader) (<-chan matchmaking.MatchMakingNotification, error) {
	data, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	playerData := model.PlayerData{}
	if err := json.Unmarshal([]byte(data), &playerData); err != nil {
		return nil, fmt.Errorf("invalid JSON received: %w", err)
	}

	return s.matchmakingService.HandlePlayerJoin(playerData), nil
}

func (s *tcpServer) handlePlayerNotifications(conn net.Conn, notifications <-chan matchmaking.MatchMakingNotification) {
	for notification := range notifications {
		json, err := json.Marshal(notification)
		if err != nil {
			slog.Error("Error marshaling JSON response", "error", err)
			continue
		}

		if _, err := conn.Write(json); err != nil {
			slog.Error("Error writing to connection", "error", err)
			return
		}

		// just to separate the notifications from each other
		if _, err := conn.Write([]byte("\n")); err != nil {
			slog.Error("Error writing to connection", "error", err)
			return
		}
	}
}

func (s *tcpServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		notifications, err := s.handlePlayerJoinRequest(reader)
		if err != nil {
			slog.Error("Error handling player join request", "error", err)
			conn.Write([]byte("closing connection... bye\n"))
			return
		}
		s.handlePlayerNotifications(conn, notifications)
	}
}

func (s *tcpServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
