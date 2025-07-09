package main

import (
	"log"

	"github.com/hrodric0/dur-impl/broadcast"
	"github.com/hrodric0/dur-impl/client"
	"github.com/hrodric0/dur-impl/server"
)

func main() {
	sequencerAddr := "localhost:8000"
	replicas := []string{"localhost:8001", "localhost:8002"}

	log.Println("[Main] Iniciando sistema DUR")
	// Inicia Sequencer
	go func() {
		log.Printf("[Sequencer] Escutando em %s", sequencerAddr)
		if err := broadcast.StartSequencer(sequencerAddr, replicas); err != nil {
			log.Fatalf("[Sequencer] erro: %v", err)
		}
	}()

	// Inicia Réplicas
	for _, addr := range replicas {
		a := addr
		go func() {
			log.Printf("[Replica %s] Inicializando", a)
			if err := server.StartReplica(a); err != nil {
				log.Fatalf("[Replica %s] erro: %v", a, err)
			}
		}()
	}

	// Exemplo de transação
	cli := client.NewTransaction("cid1", "tid1", replicas, sequencerAddr)
	log.Println("[Client cid1] Iniciando transação tid1")

	// Read
	val, err := cli.Read("x")
	if err != nil {
		log.Printf("[Client cid1] Erro em Read(x): %v", err)
	} else {
		log.Printf("[Client cid1] Read(x) -> %s", string(val))
	}

	// Write
	log.Println("[Client cid1] Write(x, new)")
	cli.Write("x", []byte("new"))

	// Commit
	log.Println("[Client cid1] Enviando Commit")
	committed, err := cli.Commit()
	if err != nil {
		log.Printf("[Client cid1] Erro em Commit: %v", err)
	} else {
		log.Printf("[Client cid1] Commit result -> %v", committed)
	}

	select {}
}
