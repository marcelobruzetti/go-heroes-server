package main

import (
	"fmt"
	"io"
	"log"
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
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if err == io.EOF {
			fmt.Println("client disconnected:", conn.RemoteAddr())
			return
		}

		if err != nil {
			fmt.Println("read error: ", err)
			return
		}

		payload := make([]byte, n)
		copy(payload, buf[:n])

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: payload,
		}

		_, err = conn.Write([]byte("thank you for your message!\n"))
		if err != nil {
			fmt.Println("write error:", err)
			return
		}
	}
}

func (s *Server) handleMessages() {
	for msg := range s.msgch {
		fmt.Printf("received message from connection (%s): %s", msg.from, string(msg.payload))
	}
}

func main() {
	server := NewServer(":3000")

	go server.handleMessages()

	log.Fatal(server.Start())
}
