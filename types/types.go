package types

// ReadRequest para leitura 1:1
type ReadRequest struct {
	Cid  string `json:"cid"`
	Item string `json:"item"`
}

// ReadReply com valor e versão
type ReadReply struct {
	Cid     string `json:"cid"`
	Item    string `json:"item"`
	Value   []byte `json:"value"`
	Version uint64 `json:"version"`
}

// ReadEntry para uso interno do client
type ReadEntry struct {
	Item    string
	Value   []byte
	Version uint64
}

// WriteEntry para uso interno do client
type WriteEntry struct {
	Item  string
	Value []byte
}

// CommitRequest enviado ao sequencer
type CommitRequest struct {
	Cid string       `json:"cid"`
	Tid string       `json:"tid"`
	Rs  []ReadEntry  `json:"rs"`
	Ws  []WriteEntry `json:"ws"`
}

// CommitDecision resposta agregada do sequencer
type CommitDecision struct {
	Cid    string `json:"cid"`
	Tid    string `json:"tid"`
	Commit bool   `json:"commit"`
}
