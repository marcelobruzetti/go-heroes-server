package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"go-heroes-server/protocol"
)

func TestWebSocketClientReceivesConnectedEvent(t *testing.T) {
	srv := NewServer(":0", ":0", nil)
	go srv.matchmakingLoop()

	httpServer := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer httpServer.Close()

	url := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test complete")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	messageType, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read connected event: %v", err)
	}
	if messageType != websocket.MessageText {
		t.Fatalf("message type = %v, want text", messageType)
	}

	var event protocol.ServerEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode connected event: %v", err)
	}
	if event.Type != "connected" {
		t.Errorf("event type = %q, want connected", event.Type)
	}
	if event.PlayerID == "" {
		t.Error("connected event did not contain a player ID")
	}
}
