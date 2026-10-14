package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-heroes-server/protocol"
	"net"

	"github.com/google/uuid"
)

// Server accepts client connections and coordinates matchmaking.
type Server struct {
	listenAddr           string
	ln                   net.Listener
	quitCh               chan struct{}
	clientConnectedCh    chan clientConnected
	clientDisconnectedCh chan clientDisconnected
}

type clientConnected struct {
	Client *Client
}

type clientDisconnected struct {
	Client *Client
}

// NewServer creates a server that listens on listenAddr.
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:           listenAddr,
		quitCh:               make(chan struct{}),
		clientConnectedCh:    make(chan clientConnected),
		clientDisconnectedCh: make(chan clientDisconnected),
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
		case event := <-s.clientConnectedCh:
			fmt.Println("matchmaking received client:", event.Client.ID)

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
			waitingClient = nil

		case event := <-s.clientDisconnectedCh:
			fmt.Println("client disconnected:", event.Client.ID)

			if waitingClient == event.Client {
				waitingClient = nil
				fmt.Println("removed waiting client:", event.Client.ID)
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

		s.clientConnectedCh <- clientConnected{
			Client: client,
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

	not json
	{"type":"ping"}
	{"type":"action","action":"dance"}
	{"type":"action","action":"attack"}
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
		case "action":
			switch msg.Action {
			case "attack", "defend", "heal":
				fmt.Println(client.ID, "- valid action")
				err := client.Send(protocol.ServerEvent{
					Type:    "success",
					Message: "valid action",
				})

				if err != nil {
					return
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
