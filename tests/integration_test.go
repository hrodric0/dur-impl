package tests

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/hrodric0/dur-impl/broadcast"
	"github.com/hrodric0/dur-impl/client"
	"github.com/hrodric0/dur-impl/server"
)

// startSystem inicializa sequencer e réplicas, aguardando até que estejam escutando
func startSystem(sequencer string, reps []string) {
	// Sequencer
	go broadcast.StartSequencer(sequencer, reps)
	// Réplicas
	for _, addr := range reps {
		go server.StartReplica(addr)
	}
	// Aguarda sequencer e réplicas estarem prontos
	// Verifica sequencer
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", sequencer)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	// Verifica réplicas
	for _, addr := range reps {
		for i := 0; i < 50; i++ {
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				conn.Close()
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	// Aguarda inicialização
	time.Sleep(100 * time.Millisecond)
}

func TestSingleTransactionCommit(t *testing.T) {
	sequencer := "localhost:9100"
	reps := []string{"localhost:9101", "localhost:9102"}
	startSystem(sequencer, reps)
	tx := client.NewTransaction("c1", "t1", reps, sequencer)
	val, err := tx.Read("x")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(val) != "init" {
		t.Fatalf("Expected 'init', got '%s'", string(val))
	}
	tx.Write("x", []byte("v1"))
	ok, err := tx.Commit()
	if err != nil || !ok {
		t.Fatalf("Expected commit success, got ok=%v err=%v", ok, err)
	}
}

func TestNonConflictingTransactions(t *testing.T) {
	sequencer := "localhost:9200"
	reps := []string{"localhost:9201", "localhost:9202"}
	startSystem(sequencer, reps)
	tx1 := client.NewTransaction("c1", "t1", reps, sequencer)
	tx2 := client.NewTransaction("c2", "t2", reps, sequencer)
	tx1.Write("a", []byte("1"))
	tx2.Write("b", []byte("2"))
	ok1, err1 := tx1.Commit()
	ok2, err2 := tx2.Commit()
	if err1 != nil || !ok1 {
		t.Fatalf("tx1 failed: ok=%v err=%v", ok1, err1)
	}
	if err2 != nil || !ok2 {
		t.Fatalf("tx2 failed: ok=%v err=%v", ok2, err2)
	}
}

func TestCommitAndAbort(t *testing.T) {
	sequencer := "localhost:9300"
	reps := []string{"localhost:9301", "localhost:9302"}
	startSystem(sequencer, reps)
	t1 := client.NewTransaction("c1", "t1", reps, sequencer)
	x1, _ := t1.Read("x")
	t2 := client.NewTransaction("c2", "t2", reps, sequencer)
	t2.Write("x", []byte("v2"))
	if ok, _ := t2.Commit(); !ok {
		t.Fatal("t2 should commit")
	}
	time.Sleep(10 * time.Millisecond)
	t1.Write("x", x1)
	if ok, _ := t1.Commit(); ok {
		t.Fatal("t1 should abort")
	}
}

// TestMultiKeyTransaction valida commit or abort em múltiplas chaves
func TestMultiKeyTransaction(t *testing.T) {
	sequencer := "localhost:9400"
	reps := []string{"localhost:9401", "localhost:9402"}
	startSystem(sequencer, reps)
	tx := client.NewTransaction("c1", "t1", reps, sequencer)
	tx.Write("a", []byte("1"))
	tx.Write("b", []byte("2"))
	ok, err := tx.Commit()
	if err != nil || !ok {
		t.Fatalf("expected multi-key commit, got ok=%v err=%v", ok, err)
	}

	// nova tx lê ambas
	tx2 := client.NewTransaction("c2", "t2", reps, sequencer)
	va, _ := tx2.Read("a")
	vb, _ := tx2.Read("b")
	if string(va) != "1" || string(vb) != "2" {
		t.Fatalf("expected reads a=1,b=2, got a=%s,b=%s", va, vb)
	}
}

// TestReadAfterCommit valida visibilidade de commit em tx posterior
func TestReadAfterCommit(t *testing.T) {
	sequencer := "localhost:9500"
	reps := []string{"localhost:9501", "localhost:9502"}
	startSystem(sequencer, reps)
	tx1 := client.NewTransaction("c1", "t1", reps, sequencer)
	tx1.Write("x", []byte("v1"))
	tx1.Commit()

	tx2 := client.NewTransaction("c2", "t2", reps, sequencer)
	val, err := tx2.Read("x")
	if err != nil || string(val) != "v1" {
		t.Fatalf("expected read-after-commit v1, got %s err=%v", val, err)
	}
}

// TestAbortThenCommitThenRead valida abort seguido de commit e leitura
func TestAbortThenCommitThenRead(t *testing.T) {
	sequencer := "localhost:9600"
	reps := []string{"localhost:9601", "localhost:9602"}
	startSystem(sequencer, reps)

	t1 := client.NewTransaction("c1", "t1", reps, sequencer)
	_, _ = t1.Read("x")

	t2 := client.NewTransaction("c2", "t2", reps, sequencer)
	t2.Write("x", []byte("v2"))
	t2.Commit()

	// força conflito: t1 aborta
	t1.Write("x", []byte("v1"))
	if ok, _ := t1.Commit(); ok {
		t.Fatal("t1 should abort")
	}

	// t3 lee após t2
	t3 := client.NewTransaction("c3", "t3", reps, sequencer)
	val, _ := t3.Read("x")
	if string(val) != "v2" {
		t.Fatalf("expected t3 read v2, got %s", val)
	}
}

// TestReadOnlyTransactions valida tx só de leitura
func TestReadOnlyTransactions(t *testing.T) {
	sequencer := "localhost:9700"
	reps := []string{"localhost:9701", "localhost:9702"}
	startSystem(sequencer, reps)

	for i := 0; i < 5; i++ {
		tx := client.NewTransaction(fmt.Sprintf("c%d", i), fmt.Sprintf("t%d", i), reps, sequencer)
		val, err := tx.Read("x")
		if err != nil || string(val) == "" {
			t.Fatalf("read-only tx failed read: val=%s err=%v", val, err)
		}
		ok, err := tx.Commit()
		if err != nil || !ok {
			t.Fatalf("read-only tx should commit, got ok=%v err=%v", ok, err)
		}
	}
}

// TestSequentialCommits valida série de commits sem conflito
func TestSequentialCommits(t *testing.T) {
	sequencer := "localhost:9800"
	reps := []string{"localhost:9801", "localhost:9802"}
	startSystem(sequencer, reps)

	for i := 0; i < 10; i++ {
		tx := client.NewTransaction(fmt.Sprintf("c%d", i), fmt.Sprintf("t%d", i), reps, sequencer)
		tx.Write("x", []byte(fmt.Sprintf("v%d", i)))
		ok, err := tx.Commit()
		if err != nil || !ok {
			t.Fatalf("sequential %d commit failed: ok=%v err=%v", i, ok, err)
		}
	}

	// verifica valor final
	tx := client.NewTransaction("cF", "tF", reps, sequencer)
	val, _ := tx.Read("x")
	if string(val) != "v9" {
		t.Fatalf("expected final value v9, got %s", val)
	}
}

// TestVariableReplicasAndClients testa múltiplas configurações de réplicas e clientes
func TestVariableReplicasAndClients(t *testing.T) {
	for idx, cfg := range []struct{ reps, clients int }{
		{1, 1}, {2, 2}, {3, 5}, {5, 10}, {10, 20},
	} {
		// Base de portas exclusivo por configuração para isolar testes
		base := 20000 + idx*100
		sequencer := fmt.Sprintf("localhost:%d", base)
		repsAddrs := make([]string, cfg.reps)
		for i := 0; i < cfg.reps; i++ {
			repsAddrs[i] = fmt.Sprintf("localhost:%d", base+1+i)
		}
		startSystem(sequencer, repsAddrs)
		t.Logf("Config %d: %d replicas, %d clients on base port %d", idx, cfg.reps, cfg.clients, base)
		var wg sync.WaitGroup
		wg.Add(cfg.clients)
		var mu sync.Mutex
		successes := 0
		for c := 0; c < cfg.clients; c++ {
			go func(id int) {
				defer wg.Done()
				tx := client.NewTransaction(
					fmt.Sprintf("c%d", id), fmt.Sprintf("t%d", id),
					repsAddrs, sequencer)
				// Leitura
				_, err := tx.Read("x")
				if err != nil {
					t.Errorf("[%d] Client %d: read error: %v", idx, id, err)
					return
				}
				// Escrita
				val := []byte(fmt.Sprintf("v%d", id))
				tx.Write("x", val)
				// Commit
				ok, err := tx.Commit()
				if err != nil {
					t.Errorf("[%d] Client %d: commit error: %v", idx, id, err)
				} else if ok {
					mu.Lock()
					successes++
					mu.Unlock()
					t.Logf("[%d] Client %d: commit succeeded", idx, id)
				} else {
					t.Logf("[%d] Client %d: commit aborted", idx, id)
				}
			}(c)
		}
		wg.Wait()
		if successes != 1 {
			t.Errorf("Config %d: expected at least 1 success, got %d", idx, successes)
		} else {
			t.Logf("Config %d: exactly 1 success as expected", idx)
		}
	}
}
