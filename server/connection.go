package server

import (
	"encoding/json"
	"go-heroes-server/protocol"
	"net"
	"sync"
)

// Client represents a connected player and its TCP connection.
type Client struct {
	ID      string
	Conn    net.Conn
	writeMu sync.Mutex
	matchID string
}

func newClient(conn net.Conn, id string) *Client {
	return &Client{
		ID:   id,
		Conn: conn,
	}
}

// Send serializes and writes an event to the client's connection.
func (c *Client) Send(event protocol.ServerEvent) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = c.Conn.Write(data)
	return err
}
