package protocol

type ClientMessage struct {
	Type   string `json:"type"`
	Action string `json:"action,omitempty"`
}

type ServerEvent struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
