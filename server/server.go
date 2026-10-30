package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-heroes-server/game"
	"go-heroes-server/protocol"
	"net"

	"github.com/google/uuid"
)

// Server accepts client connections and coordinates matchmaking.
type Server struct {
	listenAddr           string
	ln                   net.Listener
	quitCh               chan struct{}
	clientDisconnectedCh chan clientDisconnected
	clientActionCh       chan clientAction
	clientQueueCh        chan clientQueue
}

type clientQueue struct {
	Client *Client
}

type clientDisconnected struct {
	Client *Client
}

type clientAction struct {
	Client *Client
	Action game.Action
}

// NewServer creates a server that listens on listenAddr.
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:           listenAddr,
		quitCh:               make(chan struct{}),
		clientDisconnectedCh: make(chan clientDisconnected),
		clientActionCh:       make(chan clientAction),
		clientQueueCh:        make(chan clientQueue),
	}
}

// Start begins accepting client connections until the server stops.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.matchmakingLoop()
	go s.acceptLoop()

	<-s.quitCh

	return nil
}

func (s *Server) matchmakingLoop() {
	matches := make(map[string]*Match)
	var waitingClient *Client

	for {
		select {
		case event := <-s.clientQueueCh:
			fmt.Println("matchmaking received client:", event.Client.ID)

			if event.Client.matchID != "" {
				if err := event.Client.Send(protocol.ServerEvent{
					Type:    "queue_rejected",
					Message: "already in an active match",
				}); err != nil {
					fmt.Println("failed to reject queued client:", err)
				}
				continue
			}

			if waitingClient == event.Client {
				if err := event.Client.Send(protocol.ServerEvent{
					Type:    "queue_rejected",
					Message: "already waiting for an opponent",
				}); err != nil {
					fmt.Println("failed to reject queued client:", err)
				}
				continue
			}

			if waitingClient == nil {
				waitingClient = event.Client
				fmt.Println("client", waitingClient.ID, "is waiting")

				err := waitingClient.Send(protocol.ServerEvent{
					Type: "waiting",
				})

				if err != nil {
					fmt.Println("failed to notify waiting client:", err)
					waitingClient = nil
					continue
				}

				continue
			}

			first := waitingClient
			second := event.Client
			match := newMatch(first, second)
			fmt.Println("two clients joined: ", first.ID, "|", second.ID, "| MatchID: ", match.ID)

			err := first.Send(protocol.ServerEvent{
				Type:    "match_started",
				MatchID: match.ID,
			})

			if err != nil {
				fmt.Println("failed to notify first matched client:", err)
				waitingClient = second
				if err := waitingClient.Send(protocol.ServerEvent{Type: "waiting"}); err != nil {
					fmt.Println("failed to notify replacement waiting client:", err)
					waitingClient = nil
				}
				continue
			}

			err = second.Send(protocol.ServerEvent{
				Type:    "match_started",
				MatchID: match.ID,
			})

			if err != nil {
				fmt.Println("failed to notify second matched client:", err)
				waitingClient = first
				if err := waitingClient.Send(protocol.ServerEvent{Type: "waiting"}); err != nil {
					fmt.Println("failed to requeue first matched client:", err)
					waitingClient = nil
				}
				continue
			}

			first.matchID = match.ID
			second.matchID = match.ID
			matches[match.ID] = match
			broadcastGameState(match)
			waitingClient = nil

		case event := <-s.clientDisconnectedCh:
			fmt.Println("client disconnected:", event.Client.ID)

			if waitingClient == event.Client {
				waitingClient = nil
				fmt.Println("removed waiting client:", event.Client.ID)
			}

		case event := <-s.clientActionCh:
			fmt.Println(
				"received action from:", event.Client.ID,
				"| action:", event.Action,
			)

			match, found := matches[event.Client.matchID]
			if !found {
				err := event.Client.Send(protocol.ServerEvent{
					Type:    "action_rejected",
					Message: "no active match",
				})
				if err != nil {
					fmt.Println("failed to reject action:", err)
				}
				continue
			}

			if err := match.Game.ApplyAction(event.Client.ID, event.Action); err != nil {
				fmt.Println("action rejected:", err)

				if sendErr := event.Client.Send(protocol.ServerEvent{
					Type:    "action_rejected",
					Message: err.Error(),
				}); sendErr != nil {
					fmt.Println("failed to notify client:", sendErr)
				}

				continue
			}

			broadcastGameState(match)

			if match.Game.Status == game.StatusFinished {
				broadcastGameOver(match)

				delete(matches, match.ID)

				for _, player := range match.Players {
					player.matchID = ""
				}
			}
		}
	}

}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		client := newClient(conn, uuid.New().String())

		fmt.Println("new connection from:", conn.RemoteAddr(),
			" | client ID: ", client.ID)

		if err := client.Send(protocol.ServerEvent{
			Type:     "connected",
			PlayerID: client.ID,
		}); err != nil {
			fmt.Println("failed to send connected event:", err)
			conn.Close()
			continue
		}

		go s.readLoop(client)
	}
}

func (s *Server) disconnect(client *Client) {
	client.Conn.Close()
	s.clientDisconnectedCh <- clientDisconnected{
		Client: client,
	}
}

/*
msg examples:

	{"type":"queue"}
	{"type":"action","action":"attack"}
	{"type":"action","action":"defend"}
	{"type":"action","action":"heal"}
*/
func (s *Server) readLoop(client *Client) {
	defer s.disconnect(client)

	scanner := bufio.NewScanner(client.Conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		var msg protocol.ClientMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Println(client.ID, "- invalid JSON:", err)
			err := client.Send(protocol.ServerEvent{
				Type:    "error",
				Message: "invalid JSON",
			})

			if err != nil {
				return
			}

			continue
		}

		switch msg.Type {
		case "queue":
			s.clientQueueCh <- clientQueue{
				Client: client,
			}

			continue

		case "action":
			switch msg.Action {
			case "attack", "defend", "heal":
				s.clientActionCh <- clientAction{
					Client: client,
					Action: game.Action(msg.Action),
				}

				continue
			default:
				fmt.Println(client.ID, "- invalid action")
				err := client.Send(protocol.ServerEvent{
					Type:    "error",
					Message: "unknown action",
				})

				if err != nil {
					return
				}

				continue
			}
		default:
			fmt.Println(client.ID, "- invalid type")
			err := client.Send(protocol.ServerEvent{
				Type:    "error",
				Message: "unknown type",
			})

			if err != nil {
				return
			}

			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("read error: ", err)
		return
	}
}

func broadcastGameState(match *Match) {
	state := gameStateFromGame(match.Game)

	event := protocol.ServerEvent{
		Type:      "game_state",
		MatchID:   match.ID,
		GameState: state,
	}

	for _, player := range match.Players {
		if err := player.Send(event); err != nil {
			fmt.Println("failed to send game state:", err)
		}
	}
}

func broadcastGameOver(match *Match) {
	winnerID := match.Game.Players[0].ID
	if match.Game.Players[0].HP == 0 {
		winnerID = match.Game.Players[1].ID
	}

	for _, player := range match.Players {
		result := "lose"
		if player.ID == winnerID {
			result = "win"
		}

		event := protocol.ServerEvent{
			Type:    "game_over",
			MatchID: match.ID,
			Result:  result,
		}

		if err := player.Send(event); err != nil {
			fmt.Println("failed to send game over:", err)
		}
	}
}
