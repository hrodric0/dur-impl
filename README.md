# dur-impl

## Concurrency Control in Transactions with Deferred Update Replication

Implementação do protocolo **DUR (Deferred Update Replication)** em Go, com base em *Pedone & Schiper (2012)* e *Mendizabal et al. (2013)*. A solução oferece **alta concorrência local** e **consistência global** usando **difusão atômica** e **certificação de transações**.

---

## 📁 Estrutura do Projeto
```
dur/
├── go.mod                    # Definição de módulo Go
├── main.go                   # Exemplo de inicialização: sequencer + réplicas + client
├── types/
│   └── types.go              # Definição de mensagens e entradas (ReadEntry, WriteEntry, CommitRequest, etc.)
├── network/
│   └── rpc.go                # Primitivas 1:1 (Request, Send, Listen)
├── broadcast/
│   └── sequencer.go          # Implementação do sequencer (broadcast atômico centralizado)
├── client/
│   └── transaction.go        # Lógica de transação: Read, Write, Commit via sequencer
├── server/
│   └── replica.go            # Servidor réplica unificado (ReadRequest + CommitRequest)
└── tests/
    └── integration_test.go   # Testes de integração para commit, abort e concorrência
```
---

## Diagrama de Componentes (UML Component / C4 Container)
```
┌──────────────┐            ┌───────────────┐
│   Cliente    │──TCP──(1:1)──▶│  Réplica A   │
└──────────────┘            └───────────────┘
       │                           ▲   ▲
       │(1:1 Read)                 │   │(1:1 Read)
       │                           │   │
       │                           │   │
       │(1:n Commit)               │   │
       ▼                           │   │
┌──────────────┐                   │   │
│  Sequencer   │                   │   │
└──────────────┘                   │   │
       │(1:n Commit)               │   │
       ▼                           │   │
┌──────────────┐            ┌───────────────┐
│  Réplica B   │            │  Réplica C    │
└──────────────┘            └───────────────┘
```
---
## ⚙️ Componentes Principais
### 1. 🔁 Sequencer (`broadcast/sequencer.go`)
- Aguarda `CommitRequest` de clientes via TCP.
- Reenvia requisição (best-effort) a todas as réplicas na ordem recebida.
- Coleta `CommitDecision` de cada réplica e envia decisão agregada ao cliente.
- Gera logs detalhados por etapa.
---
### 2. 🧠 Réplica (`server/replica.go`)
- Listener unificado para `ReadRequest` e `CommitRequest`.
- **ReadRequest**: retorna valor e versão do `key–value store`.
- **CommitRequest**:
  - Compara `rs` com versões atuais (certificação).
  - Se houver obsolescência → **abort**.
  - Caso contrário → **commit**: aplica `ws` e incrementa versão.
- Responde com `CommitDecision` e gera logs.
---
### 3. 👨‍💻 Cliente (`client/transaction.go`)
- Estrutura `Transaction` com `rs` e `ws` locais.
- **Read**: checa `ws`; se ausente, envia `ReadRequest`.
- **Write**: grava em `ws` local.
- **Commit**: envia `CommitRequest` ao sequencer e aguarda decisão.
- Logs registram todo o fluxo.
---
### 4. 🔌 Comunicação 1:1 e 1:n (`network/rpc.go`)
- `Request`: envia JSON e espera resposta.
- `Send`: envia JSON sem esperar resposta.
- `Listen`: escuta TCP, decodifica JSON e chama o handler apropriado.
---
### 5. 🧪 Testes de Integração (`tests/integration_test.go`)
- Inicia sequencer + réplicas para cada teste.
- Testes implementados:
  - `TestSingleTransactionCommit`
  - `TestNonConflictingTransactions`
  - `TestCommitAndAbort`
---
## ✅ Pré-requisitos
- Go **1.20+**
---
## 🚀 Instruções de Uso
### Baixar dependências:
```bash
go mod tidy


Executar exemplo (inicia sequencer, réplicas e transação de demonstração):

go run main.go

Executar testes:

go test ./tests -v

O flag -v mostra logs dos componentes durante a execução dos testes.

Logs de Execução

Ao rodar go run main.go, você verá algo como:
[Main] Iniciando sistema DUR
[Sequencer] Escutando em localhost:8000
[Replica localhost:8001] Inicializando
[Replica localhost:8002] Inicializando
[Client cid1] Criando transação tid1
[Client cid1] Sending ReadRequest(item=x)
[Replica localhost:8001] Received ReadRequest cid=cid1 item=x -> value=init v0
[Client cid1] Received ReadReply: x=init (v0)
[Client cid1] Write_WS: x=new
[Client cid1] Enviando Commit
[Sequencer] Recebido CommitRequest cid=cid1 tid=tid1
[Sequencer] Enviando CommitRequest para réplica localhost:8001
[Replica localhost:8001] Received CommitRequest cid=cid1 tid=tid1
[Replica localhost:8001] Applied WS: x=v1
[Sequencer] Recebido CommitDecision from localhost:8001 -> true
... (mesma sequência para localhost:8002)
[Sequencer] Decisão agregada -> true
[Client cid1] Received CommitDecision -> true

Esses logs demonstram o fluxo otimista local (execução), difusão atômica (commit) e certificação válida.
Conclusão

Esta implementação em Go do protocolo DUR equilibra concorrência (leituras e escritas locais) e consistência (certificação via broadcast atômico). A estrutura modular facilita testes, extensões e substituição de componentes (por exemplo, outra estratégia de broadcast).
