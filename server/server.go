package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-heroes-server/protocol"
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.handleMessages()
	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		fmt.Println("new connection to the server:", conn.RemoteAddr())

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		payload := make([]byte, len(line))
		copy(payload, line)

		var msg protocol.ClientMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Println("invalid JSON:", err)
			continue
		}

		fmt.Printf("%+v\n", msg)

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: payload,
		}

		_, err := conn.Write([]byte("thank you for your message!\n"))
		if err != nil {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("read error: ", err)
		return
	}
}

func (s *Server) handleMessages() {
	for msg := range s.msgch {
		fmt.Printf("received message from connection (%s): %s", msg.from, string(msg.payload))
	}
}
