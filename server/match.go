package server

import (
	"go-heroes-server/game"

	"github.com/google/uuid"
)

// Match represents a two-player match managed by the server.
type Match struct {
	ID      string
	Players [2]*Client
	Game    *game.Game
}

func newMatch(player1 *Client, player2 *Client) *Match {
	return &Match{
		ID:      uuid.New().String(),
		Players: [2]*Client{player1, player2},
		Game:    game.NewGame(player1.ID, player2.ID),
	}
}
