package client

import (
	"log"

	"github.com/hrodric0/dur-impl/network"
	"github.com/hrodric0/dur-impl/types"
)

// Transaction mantém estado local de rs/ws
type Transaction struct {
	Cid       string
	Tid       string
	Rs        map[string]types.ReadEntry
	Ws        map[string]types.WriteEntry
	Replicas  []string
	Sequencer string
}

// NewTransaction inicializa um novo tx
func NewTransaction(cid, tid string, replicas []string, seq string) *Transaction {
	log.Printf("[Client %s] Criando transação %s", cid, tid)
	return &Transaction{Cid: cid, Tid: tid, Rs: make(map[string]types.ReadEntry), Ws: make(map[string]types.WriteEntry), Replicas: replicas, Sequencer: seq}
}

// Read usa primitiva 1:1
func (tx *Transaction) Read(item string) ([]byte, error) {
	log.Printf("[Client %s] Sending ReadRequest(item=%s)", tx.Cid, item)
	if we, ok := tx.Ws[item]; ok {
		log.Printf("[Client %s] Read from WS: %s=%s", tx.Cid, item, string(we.Value))
		return we.Value, nil
	}
	req := types.ReadRequest{Cid: tx.Cid, Item: item}
	var rep types.ReadReply
	err := network.Request(tx.Replicas[0], req, &rep)
	if err != nil {
		log.Printf("[Client %s] Read error: %v", tx.Cid, err)
		return nil, err
	}
	log.Printf("[Client %s] Received ReadReply: %s=%s (v%d)", tx.Cid, rep.Item, string(rep.Value), rep.Version)
	tx.Rs[item] = types.ReadEntry{Item: item, Value: rep.Value, Version: rep.Version}
	return rep.Value, nil
}

// Write armazena localmente
func (tx *Transaction) Write(item string, val []byte) {
	log.Printf("[Client %s] Write_WS: %s=%s", tx.Cid, item, string(val))
	tx.Ws[item] = types.WriteEntry{Item: item, Value: val}
}

// Commit faz broadcast atômico via sequencer e retorna decisão agregada
func (tx *Transaction) Commit() (bool, error) {
	log.Printf("[Client %s] Collecting rs/ws for Commit", tx.Cid)
	rs := make([]types.ReadEntry, 0, len(tx.Rs))
	for _, v := range tx.Rs {
		rs = append(rs, v)
	}
	ws := make([]types.WriteEntry, 0, len(tx.Ws))
	for _, v := range tx.Ws {
		ws = append(ws, v)
	}
	req := types.CommitRequest{Cid: tx.Cid, Tid: tx.Tid, Rs: rs, Ws: ws}
	log.Printf("[Client %s] Sending CommitRequest to Sequencer", tx.Cid)
	var dec types.CommitDecision
	err := network.Request(tx.Sequencer, req, &dec)
	if err != nil {
		log.Printf("[Client %s] Commit error: %v", tx.Cid, err)
		return false, err
	}
	log.Printf("[Client %s] Received CommitDecision -> %v", tx.Cid, dec.Commit)
	return dec.Commit, nil
}
