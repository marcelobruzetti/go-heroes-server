package server

import "github.com/google/uuid"

type Match struct {
	ID      string
	Players [2]*Client
}

func newMatch(player1 *Client, player2 *Client) *Match {
	return &Match{
		ID:      uuid.New().String(),
		Players: [2]*Client{player1, player2},
	}
}
