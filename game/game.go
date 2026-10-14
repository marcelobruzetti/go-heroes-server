// Package game defines the Go Heroes rules.
package game

import (
	"errors"
	"math/rand/v2"
)

type Status string
type Action string

const (
	StatusActive   Status = "active"
	StatusFinished Status = "finished"

	ActionAttack Action = "attack"
	ActionDefend Action = "defend"
	ActionHeal   Action = "heal"

	maxHP               = 100
	attack              = 20
	defense             = 10
	defendingMultiplier = 2
	healAmount          = 15
)

var (
	ErrGameFinished  = errors.New("game is finished")
	ErrUnknownPlayer = errors.New("unknown player")
	ErrNotYourTurn   = errors.New("not your turn")
	ErrInvalidAction = errors.New("invalid action")
)

// Game is the current game data.
type Game struct {
	Players             [2]*Player
	CurrentTurnPlayerID string
	Status              Status
}

// Player represents one combatant in a game.
type Player struct {
	ID         string
	MaxHP      int
	HP         int
	Attack     int
	Defense    int
	HealAmount int
	Defending  bool
}

// NewGame creates a game with two players.
func NewGame(firstID string, secondID string) *Game {
	currentTurnPlayerID := firstID
	if rand.N(2) == 1 {
		currentTurnPlayerID = secondID
	}

	return &Game{
		Players: [2]*Player{
			newPlayer(firstID),
			newPlayer(secondID),
		},
		CurrentTurnPlayerID: currentTurnPlayerID,
		Status:              StatusActive,
	}
}

// newPlayer creates a player with base stats.
func newPlayer(ID string) *Player {
	return &Player{
		ID:         ID,
		MaxHP:      maxHP,
		HP:         maxHP,
		Attack:     attack,
		Defense:    defense,
		HealAmount: healAmount,
		Defending:  false,
	}
}

// attackPlayer applies damage and consumes one active defense.
func attackPlayer(attacker *Player, defender *Player) {
	defense := defender.Defense
	if defender.Defending {
		defense *= defendingMultiplier
		defender.Defending = false
	}

	damage := max(attacker.Attack-defense, 0)

	defender.HP -= damage
	if defender.HP < 0 {
		defender.HP = 0
	}
}

func defendPlayer(player *Player) {
	player.Defending = true
}

func healPlayer(player *Player) {
	player.HP += player.HealAmount

	if player.HP > player.MaxHP {
		player.HP = player.MaxHP
	}
}

func (g *Game) ApplyAction(playerID string, action Action) error {
	if g.Status == StatusFinished {
		return ErrGameFinished
	}

	if playerID != g.Players[0].ID && playerID != g.Players[1].ID {
		return ErrUnknownPlayer
	}

	if playerID != g.CurrentTurnPlayerID {
		return ErrNotYourTurn
	}

	attacker := g.Players[0]
	defender := g.Players[1]

	if attacker.ID != playerID {
		attacker = g.Players[1]
		defender = g.Players[0]
	}

	switch action {
	case ActionAttack:
		attackPlayer(attacker, defender)
	case ActionDefend:
		defendPlayer(attacker)
	case ActionHeal:
		healPlayer(attacker)
	default:
		return ErrInvalidAction
	}

	if defender.HP == 0 {
		g.Status = StatusFinished
	} else {
		g.CurrentTurnPlayerID = defender.ID
	}

	return nil
}
