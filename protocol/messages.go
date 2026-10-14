package protocol

// ClientMessage is a command sent from a client to the server.
type ClientMessage struct {
	Type   string `json:"type"`
	Action string `json:"action,omitempty"`
}

// ServerEvent is a message sent from the server to a client.
type ServerEvent struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
	MatchID string `json:"match_id,omitempty"`
}
