package broadcast

import (
	"encoding/json"
	"log"
	"net"

	"github.com/hrodric0/dur-impl/types"
)

// StartSequencer implementa broadcast atômico com ordenação FIFO garantida.
func StartSequencer(listenAddr string, replicaAddrs []string) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	// Canal para requisições recebidas
	type reqConn struct {
		req  types.CommitRequest
		conn net.Conn
	}
	ch := make(chan reqConn, 100)

	// Processador sequencial de commits
	go func() {
		for rc := range ch {
			r := rc.req
			timeSpent := log.Printf // alias para evitar import cycl
			timeSpent("[Sequencer] Processando CommitRequest cid=%s tid=%s", r.Cid, r.Tid)
			agg := true
			for _, addr := range replicaAddrs {
				log.Printf("[Sequencer] Enviando a réplica %s", addr)
				conn2, err := net.Dial("tcp", addr)
				if err != nil {
					log.Printf("[Sequencer] falha conectar %s: %v", addr, err)
					agg = false
					continue
				}
				json.NewEncoder(conn2).Encode(r)
				var dec types.CommitDecision
				json.NewDecoder(conn2).Decode(&dec)
				log.Printf("[Sequencer] Decisão da réplica %s -> %v", addr, dec.Commit)
				if !dec.Commit {
					agg = false
				}
				conn2.Close()
			}
			// Retorna decisão ao cliente
			out := types.CommitDecision{Cid: r.Cid, Tid: r.Tid, Commit: agg}
			json.NewEncoder(rc.conn).Encode(out)
			rc.conn.Close()
		}
	}()

	// Aceita conexões e envia para canal
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		var req types.CommitRequest
		if err := json.NewDecoder(conn).Decode(&req); err != nil {
			conn.Close()
			continue
		}
		// Enfileira para processamento ordenado
		ch <- reqConn{req: req, conn: conn}
	}
}
