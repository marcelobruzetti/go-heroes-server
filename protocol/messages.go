// Package protocol defines messages exchanged by Go Heroes clients and server.
package protocol

// ClientMessage is a command sent from a client to the server.
type ClientMessage struct {
	Type   string `json:"type"`
	Action string `json:"action,omitempty"`
}

// ServerEvent is a message sent from the server to a client.
type ServerEvent struct {
	Type      string     `json:"type"`
	Message   string     `json:"message,omitempty"`
	MatchID   string     `json:"match_id,omitempty"`
	PlayerID  string     `json:"player_id,omitempty"`
	GameState *GameState `json:"game_state,omitempty"`
	Result    string     `json:"result,omitempty"`
}

type PlayerState struct {
	ID    string `json:"id"`
	HP    int    `json:"hp"`
	MaxHP int    `json:"max_hp"`
}

type GameState struct {
	Players             [2]PlayerState `json:"players"`
	CurrentTurnPlayerID string         `json:"current_turn_player_id"`
	Status              string         `json:"status"`
}
