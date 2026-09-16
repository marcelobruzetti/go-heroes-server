# Go Heroes Server — Development Plan

## 1. Goal

Build a small authoritative multiplayer server in **Go** for Go Heroes.

The MVP exists primarily to learn Go networking and concurrency. Gameplay complexity is intentionally constrained.

## 2. MVP boundary

The server must support:

1. TCP connections from Unity clients.
2. Multiple clients concurrently.
3. Automatic two-player matchmaking.
4. A turn-based match with Attack, Defend, and Heal.
5. Server-authoritative validation and state.
6. Victory/defeat.
7. Returning finished players to matchmaking when they choose **Play Again**.
8. Safe handling of invalid messages and disconnects.

Not part of the MVP:

- Accounts/authentication.
- PostgreSQL.
- Rankings.
- Inventory/progression.
- Lobby or room list.
- Browser client.
- WebSocket.
- gRPC.
- True reconnect/session recovery.

---

## Phase 1 — Minimal TCP server

### Objective

Learn the basic Go TCP lifecycle before adding game concepts.

### Tasks

- Initialize the Go module.
- Start a TCP listener.
- Accept client connections.
- Handle multiple connections concurrently.
- Read framed messages.
- Write responses.
- Detect disconnected clients.
- Implement graceful shutdown.

### Done when

Two terminal clients can connect simultaneously and exchange test messages without interfering with each other.

---

## Phase 2 — Application protocol

### Objective

Create the smallest reliable message contract between client and server.

### Initial approach

Use newline-delimited JSON over TCP.

```json
{"type":"action","action":"attack"}
```

Possible server events include:

- `waiting`
- `match_started`
- `game_state`
- `action_rejected`
- `game_over`
- `error`

Do not over-design the protocol before the gameplay flow needs each message.

### Tasks

- Define client message DTOs.
- Define server event DTOs.
- Implement message framing.
- Decode/encode JSON.
- Reject malformed JSON.
- Reject unknown message/action types.
- Keep game/domain models independent from transport DTOs where useful.

### Done when

A terminal client can send structured commands and receive structured events reliably.

---

## Phase 3 — Matchmaking

### Objective

Automatically pair players.

### Flow

```text
Client connects
      │
      ▼
Waiting
      │
      │ another waiting player exists
      ▼
Create match
      │
      ▼
Send match_started to both
```

### Tasks

- Create a waiting queue.
- Make shared matchmaking state concurrency-safe.
- Assign player IDs.
- Generate match IDs.
- Associate clients with their active match.
- Remove disconnected waiting players.

### Done when

Four clients can connect and become two independent matches.

---

## Phase 4 — Pure game rules

### Objective

Implement the duel independently from TCP.

### Initial rules

Keep the numbers configurable/simple while developing.

- Two players.
- Fixed maximum/initial HP.
- One active player per turn.
- `Attack`.
- `Defend`.
- `Heal`.
- Out-of-turn actions are invalid.
- HP cannot exceed maximum HP.
- A player at `0 HP` loses.
- Finished matches reject gameplay actions.

### Suggested model

```text
Game
├── ID
├── Players[2]
├── CurrentTurn
└── Status

Player
├── ID
├── HP
└── Defending
```

### Tests

Cover at least:

- Attack.
- Defend.
- Heal.
- Heal at maximum HP.
- Invalid turn.
- Invalid action.
- Defense interaction.
- Defeat.
- Finished game rejecting actions.

### Done when

A complete match can be simulated entirely in Go tests without opening a socket.

---

## Phase 5 — Connect game + TCP

### Objective

Turn the standalone networking and game rules into a playable server.

### Tasks

- Create a game when matchmaking pairs two players.
- Decide/sort the starting player.
- Translate network actions into game commands.
- Validate actions against the authoritative game state.
- Broadcast updated state to both players.
- Send `game_over` with each client's result.
- Remove finished matches from active server state.

### Important rule

Never trust client-supplied HP, turn, damage, winner, or match state.

The client says:

```text
"I want to attack."
```

The server decides what that means.

### Done when

Two terminal clients can play a complete match through TCP.

---

## Phase 6 — Unity integration contract

### Objective

Support the exact UI flow implemented by the Go Heroes Unity repository.

### Expected client states

```text
Play
  ↓
Waiting for opponent
  ↓
Playing
  ↓
Win / Lose
  ↓
Play Again
  ↓
Waiting for opponent
```

`Play Again` means entering matchmaking for a new match. It is **not** true network reconnection/session recovery.

### Tasks

- Send a waiting state after the player enters matchmaking.
- Send initial state to both clients when a match starts.
- Keep both clients synchronized after every valid action.
- Return useful errors for rejected actions.
- Send explicit win/loss information.
- Support re-queueing after a finished match.

### Done when

Two Unity clients can complete this full flow without server restarts.

---

## Phase 7 — Robustness

### Test scenarios

- Disconnect while waiting.
- Disconnect during a match.
- Malformed JSON.
- Unknown message type.
- Unknown action.
- Action outside the player's turn.
- Duplicate/rapid actions.
- Multiple simultaneous matches.
- `Play Again` after a finished match.
- Server shutdown with active connections.

### Quality

- `gofmt`.
- `go vet`.
- Automated unit tests.
- Race detector where practical (`go test -race ./...`).
- Clear logging without leaking transport concerns into game rules.
- README setup/run instructions.

### MVP definition of done

The server MVP is complete when:

1. It accepts concurrent TCP clients.
2. Automatic matchmaking works.
3. Matches are isolated from one another.
4. Attack, Defend, and Heal are authoritative server actions.
5. Turns are enforced.
6. Victory/defeat is detected.
7. Both Unity clients remain synchronized.
8. Players can choose Play Again and return to matchmaking.
9. Invalid input/disconnects do not crash the server.
10. Core rules have automated tests.

---

# Future evolution

## Lobby and rooms

Replace or complement automatic matchmaking with:

- Create room.
- List rooms.
- Join/leave room.
- Room name/code.
- Waiting/playing status.
- Optional rematch with the same opponent.

## Multiple transports

Only after TCP works, extract the boundaries needed to support another transport.

### WebSocket

Use the same game core with a browser-compatible persistent connection.

```text
Unity ───── TCP ───────┐
                       ├── Game core
Browser ─ WebSocket ───┘
```

### gRPC

Experiment with Protocol Buffers and streaming where it provides learning value.

The purpose is to compare architectural and operational trade-offs, not to replace TCP automatically.

## HTTP API

Potential supporting endpoints:

- Health.
- Server info.
- Room list.
- Match history.

## Persistence

PostgreSQL may later store:

- Player profiles.
- Match history.
- Win/loss statistics.

Do not persist active match state in the MVP.

## True reconnect

A later version may introduce player/session tokens so a dropped client can recover an in-progress match. This is intentionally separate from the MVP's **Play Again** flow.