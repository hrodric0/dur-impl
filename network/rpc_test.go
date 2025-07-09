package network

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestRequestAndListen(t *testing.T) {
	// dummy listener
	addr := "localhost:0"
	handler := func(raw []byte, conn net.Conn) {
		var msg map[string]string
		json.Unmarshal(raw, &msg)
		resp := map[string]string{"echo": msg["ping"]}
		json.NewEncoder(conn).Encode(resp)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}
	go Listen(ln.Addr().String(), handler)
	time.Sleep(10 * time.Millisecond)

	// perform Request
	req := map[string]string{"ping": "pong"}
	var resp map[string]string
	if err := Request(ln.Addr().String(), req, &resp); err != nil {
		t.Fatalf("Request error: %v", err)
	}
	if resp["echo"] != "pong" {
		t.Errorf("Expected echo=pong, got %v", resp)
	}
}
