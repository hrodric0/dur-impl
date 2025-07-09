package server

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/hrodric0/dur-impl/types"
)

func TestReplicaCertification(t *testing.T) {
	// start replica
	ln, _ := net.Listen("tcp", "localhost:0")
	go StartReplica(ln.Addr().String())
	time.Sleep(10 * time.Millisecond)

	// send commit with stale rs
	req := types.CommitRequest{Cid: "c", Tid: "t", Rs: []types.ReadEntry{{Item: "x", Version: 999}}, Ws: nil}
	conn, _ := net.Dial("tcp", ln.Addr().String())
	json.NewEncoder(conn).Encode(req)
	var dec types.CommitDecision
	json.NewDecoder(conn).Decode(&dec)
	conn.Close()
	if dec.Commit {
		t.Errorf("Expected abort for stale read, got commit")
	}
}
