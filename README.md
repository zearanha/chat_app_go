# Chat em Tempo Real com WebSocket em Go

Servidor de chat em tempo real construído em Go puro (+ Gorilla WebSocket), usando o padrão **Hub** para gerenciar conexões concorrentes sem race conditions — via channels, não mutexes.

## ✨ Motivação

Este projeto foi criado como exercício prático de concorrência em Go, explorando o princípio:

> "Don't communicate by sharing memory; share memory by communicating."

Em vez de proteger um mapa de clientes com `sync.Mutex`, toda a lógica de estado passa por uma única goroutine (`Hub.run()`), que recebe eventos via channels.

## 🏗️ Arquitetura

```
Cliente A (WS) ─┐
Cliente B (WS) ─┼──> Hub (goroutine central) ──> Broadcast para todos os clientes
Cliente C (WS) ─┘
```

- **Hub**: centraliza registro/remoção de clientes e broadcast de mensagens
- **Client**: cada conexão roda duas goroutines dedicadas —`readPump` (lê do WebSocket) e `writePump` (escreve no WebSocket)
- Comunicação entre Hub e Clients é feita 100% via channels (`register`, `unregister`, `broadcast`, `send`)

## 📂 Estrutura do projeto

```
chat-app/
├── go.mod
├── main.go       # inicializa o servidor HTTP e o Hub
├── hub.go        # lógica central de gerenciamento de clientes
├── client.go     # readPump / writePump de cada conexão
└── handlers.go   # upgrade de HTTP para WebSocket
```

## 🚀 Como rodar

```bash
git clone <repo>
cd chat-app
go mod tidy
go run .
```

O servidor sobe em `http://localhost:8080`, com o endpoint WebSocket em `ws://localhost:8080/ws`.

## 🧪 Como testar

**Via console do navegador:**

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onmessage = (event) => console.log("Recebido:", event.data);
ws.send("oi do navegador");
```

Abra múltiplas abas para simular vários clientes trocando mensagens.

**Via `wscat`:**

```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws
```

## ✅ Validação de concorrência

O projeto foi testado com a flag `-race` do Go, simulando múltiplos clientes enviando mensagens simultaneamente:

```bash
go run -race .
```

Nenhuma race condition é esperada, já que o `map[*Client]bool` só é acessado dentro da goroutine `Hub.run()`.

## 🛣️ Roadmap (próximas fases)

- [ ] Persistência de mensagens (PostgreSQL)
- [ ] Autenticação de usuários (JWT)
- [ ] Salas de chat (rooms) separadas
- [ ] Escala horizontal com Redis Pub/Sub
- [ ] Indicador de "digitando..." e presença online

## 🛠️ Tecnologias

- [Go](https://go.dev/)
- [gorilla/websocket](https://github.com/gorilla/websocket)

## 📚 Aprendizados

- Uso de goroutines e channels para gerenciar estado concorrente sem locks explícitos
- Diferença entre `readPump` e `writePump` rodando em paralelo por conexão
- Handshake HTTP → WebSocket e por que ele não funciona direto pela barra de endereço do navegador
- Broadcast eficiente para múltiplos clientes com tratamento de clientes lentos/desconectados
