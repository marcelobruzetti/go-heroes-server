package server

import (
	"go-heroes-server/game"
	"go-heroes-server/protocol"
)

func gameStateFromGame(g *game.Game) *protocol.GameState {
	return &protocol.GameState{
		Players: [2]protocol.PlayerState{
			{
				ID:    g.Players[0].ID,
				HP:    g.Players[0].HP,
				MaxHP: g.Players[0].MaxHP,
			},
			{
				ID:    g.Players[1].ID,
				HP:    g.Players[1].HP,
				MaxHP: g.Players[1].MaxHP,
			},
		},
		CurrentTurnPlayerID: g.CurrentTurnPlayerID,
		Status:              string(g.Status),
	}
}
