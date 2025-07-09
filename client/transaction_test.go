package client

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/hrodric0/dur-impl/types"
)

func TestReadPopulatesRs(t *testing.T) {
	// start fake replica
	ln, _ := net.Listen("tcp", "localhost:0")
	go func() {
		for {
			conn, _ := ln.Accept()
			var req types.ReadRequest
			json.NewDecoder(conn).Decode(&req)
			// respond with fixed value
			rep := types.ReadReply{Cid: req.Cid, Item: req.Item, Value: []byte("val"), Version: 5}
			json.NewEncoder(conn).Encode(rep)
			conn.Close()
		}
	}()
	time.Sleep(10 * time.Millisecond)

	tx := NewTransaction("c1", "t1", []string{ln.Addr().String()}, "")
	v, err := tx.Read("x")
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if string(v) != "val" {
		t.Errorf("Expected val, got %s", v)
	}
	entry, ok := tx.Rs["x"]
	if !ok || entry.Version != 5 {
		t.Errorf("Rs not populated correctly: %+v", tx.Rs)
	}
}
