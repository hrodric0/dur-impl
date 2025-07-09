package server

import (
	"encoding/json"
	"log"
	"net"

	"github.com/hrodric0/dur-impl/network"
	"github.com/hrodric0/dur-impl/types"
)

// VersionedValue armazena valor e versão de cada chave
type VersionedValue struct {
	Value   []byte
	Version uint64
}

// Replica mantém estado do KV e contador de versões
type Replica struct {
	Addr          string
	Db            map[string]VersionedValue
	LastCommitted uint64
}

// StartReplica inicia listener unificado para Read/Commit
func StartReplica(addr string) error {
	log.Printf("[Replica %s] Escutando...", addr)
	rep := &Replica{Addr: addr, Db: map[string]VersionedValue{"x": {Value: []byte("init"), Version: 0}}, LastCommitted: 0}
	handler := func(raw []byte, c net.Conn) {
		var probe map[string]json.RawMessage
		json.Unmarshal(raw, &probe)
		if _, isCommit := probe["rs"]; isCommit {
			var req types.CommitRequest
			json.Unmarshal(raw, &req)
			log.Printf("[Replica %s] Received CommitRequest cid=%s tid=%s", addr, req.Cid, req.Tid)
			// certificação
			abort := false
			for _, re := range req.Rs {
				if vv, ok := rep.Db[re.Item]; ok && vv.Version > re.Version {
					abort = true
					break
				}
			}
			if abort {
				log.Printf("[Replica %s] DECISION abort (rs stale)", addr)
			} else {
				rep.LastCommitted++
				for _, we := range req.Ws {
					rep.Db[we.Item] = VersionedValue{Value: we.Value, Version: rep.LastCommitted}
					log.Printf("[Replica %s] Applied WS: %s=v%d", addr, we.Item, rep.LastCommitted)
				}
				log.Printf("[Replica %s] DECISION commit", addr)
			}
			decision := types.CommitDecision{Cid: req.Cid, Tid: req.Tid, Commit: !abort}
			json.NewEncoder(c).Encode(decision)
		} else {
			var req types.ReadRequest
			json.Unmarshal(raw, &req)
			vv := rep.Db[req.Item]
			log.Printf("[Replica %s] Received ReadRequest cid=%s item=%s -> value=%s v%d", addr, req.Cid, req.Item, string(vv.Value), vv.Version)
			repMsg := types.ReadReply{Cid: req.Cid, Item: req.Item, Value: vv.Value, Version: vv.Version}
			json.NewEncoder(c).Encode(repMsg)
		}
	}
	return network.Listen(rep.Addr, handler)
}
