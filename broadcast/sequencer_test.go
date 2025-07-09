package broadcast

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/hrodric0/dur-impl/types"
)

func TestSequencerFIFO(t *testing.T) {
	// dummy replica
	ln, _ := net.Listen("tcp", "localhost:0")
	go func() {
		for {
			conn, _ := ln.Accept()
			var req types.CommitRequest
			json.NewDecoder(conn).Decode(&req)
			json.NewEncoder(conn).Encode(types.CommitDecision{Cid: req.Cid, Tid: req.Tid, Commit: true})
			conn.Close()
		}
	}()
	time.Sleep(10 * time.Millisecond)
	// start sequencer
	seqLn, _ := net.Listen("tcp", "localhost:0")
	go StartSequencer(seqLn.Addr().String(), []string{ln.Addr().String()})
	time.Sleep(10 * time.Millisecond)

	// send two requests
	c1, _ := net.Dial("tcp", seqLn.Addr().String())
	json.NewEncoder(c1).Encode(types.CommitRequest{Cid: "1", Tid: "t1"})
	c2, _ := net.Dial("tcp", seqLn.Addr().String())
	json.NewEncoder(c2).Encode(types.CommitRequest{Cid: "2", Tid: "t2"})

	var d1, d2 types.CommitDecision
	json.NewDecoder(c1).Decode(&d1)
	json.NewDecoder(c2).Decode(&d2)
	if d1.Tid != "t1" || d2.Tid != "t2" {
		t.Errorf("Expected FIFO order, got %v then %v", d1.Tid, d2.Tid)
	}
}
