package game

import (
	"errors"
	"testing"
)

func TestNewGameInitializesPlayers(t *testing.T) {
	game := NewGame("player-1", "player-2")

	if game.Status != StatusActive {
		t.Fatalf("expected status %q, got %q", StatusActive, game.Status)
	}

	if game.Players[0].ID != "player-1" {
		t.Fatalf("expected first player ID %q, got %q", "player-1", game.Players[0].ID)
	}

	if game.Players[1].ID != "player-2" {
		t.Fatalf("expected second player ID %q, got %q", "player-2", game.Players[1].ID)
	}

	for _, player := range game.Players {
		if player.HP != maxHP {
			t.Fatalf("expected player %q HP to be %d, got %d", player.ID, maxHP, player.HP)
		}

		if player.MaxHP != maxHP {
			t.Fatalf("expected player %q max HP to be %d, got %d", player.ID, maxHP, player.MaxHP)
		}

		if player.Defending {
			t.Fatalf("expected player %q not to start defending", player.ID)
		}
	}
}

func TestNewGameChoosesOneOfThePlayersToStart(t *testing.T) {
	game := NewGame("player-1", "player-2")

	currentTurn := game.CurrentTurnPlayerID
	if currentTurn != "player-1" && currentTurn != "player-2" {
		t.Fatalf("expected current turn to belong to a game player, got %q", currentTurn)
	}
}

func TestAttackPlayerReducesDefenderHP(t *testing.T) {
	attacker := newPlayer("attacker")
	defender := newPlayer("defender")

	attackPlayer(attacker, defender)

	wantHP := maxHP - (attacker.Attack - defender.Defense)
	if defender.HP != wantHP {
		t.Fatalf("expected defender HP to be %d, got %d", wantHP, defender.HP)
	}

	if defender.Defending {
		t.Fatal("expected non-defending player to remain non-defending")
	}
}

func TestAttackPlayerConsumesDefense(t *testing.T) {
	attacker := newPlayer("attacker")
	defender := newPlayer("defender")
	defender.Defending = true

	attackPlayer(attacker, defender)

	if defender.HP != maxHP {
		t.Fatalf("expected defending player HP to remain %d, got %d", maxHP, defender.HP)
	}

	if defender.Defending {
		t.Fatal("expected defense to be consumed after an attack")
	}
}

func TestApplyActionAttackDamagesOpponentAndChangesTurn(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	hpBefore := game.Players[1].HP

	err := game.ApplyAction("player-1", ActionAttack)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wantHP := hpBefore - (game.Players[0].Attack - game.Players[1].Defense)

	if game.Players[1].HP != wantHP {
		t.Fatalf("expected player-2 HP to be %d, got %d", wantHP, game.Players[1].HP)
	}

	if game.CurrentTurnPlayerID != "player-2" {
		t.Fatalf("expected player-2 turn, got %q", game.CurrentTurnPlayerID)
	}

	if game.Status != StatusActive {
		t.Fatalf("expected game to remain active, got %q", game.Status)
	}
}

func TestApplyActionDefendSetsDefendingAndChangesTurn(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	err := game.ApplyAction("player-1", ActionDefend)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wantDefending := true

	if game.Players[0].Defending != wantDefending {
		t.Fatalf("expected player-1 Defending to be %t, got %t", wantDefending, game.Players[0].Defending)
	}

	if game.CurrentTurnPlayerID != "player-2" {
		t.Fatalf("expected player-2 turn, got %q", game.CurrentTurnPlayerID)
	}

	if game.Status != StatusActive {
		t.Fatalf("expected game to remain active, got %q", game.Status)
	}
}

func TestApplyActionHealRecoversHPAndChangesTurn(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	hpBefore := 50
	game.Players[0].HP = hpBefore

	err := game.ApplyAction("player-1", ActionHeal)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wantHP := hpBefore + game.Players[0].HealAmount
	if wantHP > game.Players[0].MaxHP {
		wantHP = game.Players[0].MaxHP
	}

	if game.Players[0].HP != wantHP {
		t.Fatalf("expected player-1 HP to be %d, got %d", wantHP, game.Players[0].HP)
	}

	if game.CurrentTurnPlayerID != "player-2" {
		t.Fatalf("expected player-2 turn, got %q", game.CurrentTurnPlayerID)
	}

	if game.Status != StatusActive {
		t.Fatalf("expected game to remain active, got %q", game.Status)
	}
}

func TestApplyActionAttackDefeatsOpponentAndFinishTheGame(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	game.Players[1].HP = 10

	err := game.ApplyAction("player-1", ActionAttack)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wantHP := 0

	if game.Players[1].HP != wantHP {
		t.Fatalf("expected player-2 HP to be 0, got %d", game.Players[1].HP)
	}

	if game.Status != StatusFinished {
		t.Fatalf("expected game status %q, got %q", StatusFinished, game.Status)
	}
}

func TestApplyActionRejectsActionAfterGameFinished(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"
	game.Players[1].HP = 10

	if err := game.ApplyAction("player-1", ActionAttack); err != nil {
		t.Fatalf("expected killing attack to succeed, got %v", err)
	}

	hpBefore := game.Players[0].HP
	turnBefore := game.CurrentTurnPlayerID

	err := game.ApplyAction("player-2", ActionHeal)

	if !errors.Is(err, ErrGameFinished) {
		t.Fatalf("expected ErrGameFinished, got %v", err)
	}

	if game.Players[0].HP != hpBefore {
		t.Fatalf("expected rejected action not to change player-1 HP, got %d", game.Players[0].HP)
	}

	if game.CurrentTurnPlayerID != turnBefore {
		t.Fatalf("expected rejected action not to change turn, got %q", game.CurrentTurnPlayerID)
	}
}

func TestApplyActionRejectsInvalidActionWithoutChangingGame(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	player1HPBefore := game.Players[0].HP
	player2HPBefore := game.Players[1].HP
	turnBefore := game.CurrentTurnPlayerID

	err := game.ApplyAction("player-1", Action("dance"))

	if !errors.Is(err, ErrInvalidAction) {
		t.Fatalf("expected ErrInvalidAction, got %v", err)
	}

	if game.Players[0].HP != player1HPBefore || game.Players[1].HP != player2HPBefore {
		t.Fatal("expected invalid action not to change HP")
	}

	if game.CurrentTurnPlayerID != turnBefore {
		t.Fatalf("expected invalid action not to change turn, got %q", game.CurrentTurnPlayerID)
	}
}

func TestApplyActionHealDoesNotExceedMaxHP(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"
	game.Players[0].HP = 95

	err := game.ApplyAction("player-1", ActionHeal)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if game.Players[0].HP != game.Players[0].MaxHP {
		t.Fatalf(
			"expected player-1 HP to be capped at %d, got %d",
			game.Players[0].MaxHP,
			game.Players[0].HP,
		)
	}

	if game.CurrentTurnPlayerID != "player-2" {
		t.Fatalf("expected player-2 turn, got %q", game.CurrentTurnPlayerID)
	}
}

func TestApplyActionRejectsOutOfTurnPlayer(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	player1HPBefore := game.Players[0].HP
	player2HPBefore := game.Players[1].HP
	turnBefore := game.CurrentTurnPlayerID

	err := game.ApplyAction("player-2", ActionAttack)

	if !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}

	if game.Players[0].HP != player1HPBefore || game.Players[1].HP != player2HPBefore {
		t.Fatal("expected rejected action not to change HP")
	}

	if game.CurrentTurnPlayerID != turnBefore {
		t.Fatalf("expected rejected action not to change turn, got %q", game.CurrentTurnPlayerID)
	}
}

func TestApplyActionRejectsUnknownPlayer(t *testing.T) {
	game := NewGame("player-1", "player-2")
	game.CurrentTurnPlayerID = "player-1"

	player1HPBefore := game.Players[0].HP
	player2HPBefore := game.Players[1].HP
	turnBefore := game.CurrentTurnPlayerID

	err := game.ApplyAction("player-3", ActionAttack)

	if !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("expected ErrUnknownPlayer, got %v", err)
	}

	if game.Players[0].HP != player1HPBefore || game.Players[1].HP != player2HPBefore {
		t.Fatal("expected rejected action not to change HP")
	}

	if game.CurrentTurnPlayerID != turnBefore {
		t.Fatalf("expected rejected action not to change turn, got %q", game.CurrentTurnPlayerID)
	}
}
