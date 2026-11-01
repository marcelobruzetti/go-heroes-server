package server

import (
	"encoding/json"
	"go-heroes-server/protocol"
	"io"
	"sync"
)

// Client represents a connected player independently of its network transport.
type Client struct {
	ID        string
	reader    io.Reader
	send      func([]byte) error
	close     func() error
	writeMu   sync.Mutex
	closeOnce sync.Once
	matchID   string
}

func newClient(id string, reader io.Reader, send func([]byte) error, close func() error) *Client {
	return &Client{
		ID:     id,
		reader: reader,
		send:   send,
		close:  close,
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

	return c.send(data)
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		_ = c.close()
	})
}
