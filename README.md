# Chat em Tempo Real com WebSocket em Go

Servidor de chat em tempo real construído em Go puro (+ Gorilla WebSocket), usando o padrão **Hub** para gerenciar conexões concorrentes sem race conditions — via channels, não mutexes. Persistência com MongoDB, autenticação via JWT e suporte a múltiplas salas.

## ✨ Motivação

Este projeto foi criado como exercício prático de concorrência em Go, explorando o princípio:

> "Don't communicate by sharing memory; share memory by communicating."

Em vez de proteger um mapa de clientes com `sync.Mutex`, toda a lógica de estado passa por uma única goroutine (`Hub.run()`), que recebe eventos via channels.

## 🏗️ Arquitetura

```
Cliente A (WS, sala X) ─┐
Cliente B (WS, sala X) ─┼──> Hub (goroutine central) ──> Broadcast filtrado por sala
Cliente C (WS, sala Y) ─┘                                        │
                                                             MongoDB (mensagens + usuários)
```

- **Hub**: centraliza registro/remoção de clientes e broadcast de mensagens, filtrando por sala (`room`)
- **Client**: cada conexão roda duas goroutines dedicadas — `readPump` (lê do WebSocket) e `writePump` (escreve no WebSocket)
- **Store**: camada de acesso ao MongoDB — mensagens (com histórico por sala) e usuários (com senha em hash)
- **Auth**: registro/login via REST, geração e validação de JWT; conexão WebSocket exige token válido
- Comunicação entre Hub e Clients é feita 100% via channels (`register`, `unregister`, `broadcast`, `send`)

## 📂 Estrutura do projeto

```
chat-app/
├── go.mod
├── cmd/
│   └── server/
│       └── main.go          # bootstrap do servidor HTTP, Store e Hub
└── internal/
    ├── auth/                # hash de senha e geração/validação de JWT
    ├── chat/                # Hub, Client e fluxo WebSocket
    ├── httpapi/             # rotas HTTP, auth handlers e handler WebSocket
    └── store/               # acesso ao MongoDB, mensagens e usuários
```

## 🚀 Como rodar

**Pré-requisitos:** Go instalado, MongoDB rodando localmente (`mongodb://localhost:27017`).

```bash
git clone https://github.com/zearanha/chat_app_go
cd chat_app_go
go mod tidy
go run ./cmd/server
```

O servidor sobe em `http://localhost:8080`.

## 📡 Endpoints

| Método | Rota        | Descrição                                    |
|--------|-------------|-----------------------------------------------|
| POST   | `/register` | Cria um novo usuário e retorna um token JWT   |
| POST   | `/login`    | Autentica um usuário e retorna um token JWT   |
| GET    | `/rooms`    | Lista as salas com mensagens existentes       |
| GET    | `/ws`       | Upgrade para WebSocket (requer `token` e `room` via query string) |

## 🧪 Como testar

**1. Registrar usuário:**

```powershell
Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body '{"username":"jose","password":"senha123"}' -ContentType "application/json"
```

**2. Conectar no WebSocket com token e sala (console do navegador):**

```javascript
const token = "TOKEN_RECEBIDO_NO_REGISTRO_OU_LOGIN";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}&room=golang`);

ws.onmessage = (event) => console.log("Recebido:", event.data);
ws.send("primeira mensagem na sala golang");
```

Abra múltiplas abas com salas diferentes (`room=golang` vs `room=geral`) para confirmar o isolamento entre elas.

## ✅ Validação de concorrência

O projeto foi testado com a flag `-race` do Go, simulando múltiplos clientes enviando mensagens simultaneamente:

```bash
go run -race ./cmd/server
```

Nenhuma race condition é esperada, já que o `map[*Client]bool` só é acessado dentro da goroutine `Hub.run()`.

## 🛣️ Roadmap

- [x] Comunicação em tempo real via WebSocket (padrão Hub)
- [x] Persistência de mensagens (MongoDB)
- [x] Autenticação de usuários (JWT)
- [x] Salas de chat (rooms) separadas com histórico próprio
- [ ] Escala horizontal com Redis Pub/Sub *(fora do escopo por ora)*

## 🛠️ Tecnologias

- [Go](https://go.dev/)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

## 📚 Aprendizados

- Uso de goroutines e channels para gerenciar estado concorrente sem locks explícitos
- Diferença entre `readPump` e `writePump` rodando em paralelo por conexão
- Handshake HTTP → WebSocket e por que ele não funciona direto pela barra de endereço do navegador
- Broadcast eficiente para múltiplos clientes com tratamento de clientes lentos/desconectados
- Persistência de dados com MongoDB (índices, filtros, ordenação)
- Autenticação stateless com JWT e proteção de conexões WebSocket via token
- Isolamento de mensagens por sala reaproveitando um único Hub central
