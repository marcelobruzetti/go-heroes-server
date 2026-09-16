# Go Heroes Server

The authoritative multiplayer server for **Go Heroes**, a deliberately small two-player turn-based game built as a learning project for Go networking and multiplayer server architecture.

The Unity client lives in a separate repository. This repository focuses on the server.

## The game

Two players are automatically matched into a duel. Each player has HP and, on their turn, chooses one action:

- **Attack** — damage the opponent.
- **Defend** — reduce or block incoming damage according to the game rules.
- **Heal** — restore HP up to the maximum.

The first player to reach `0 HP` loses.

## Why this project exists

The gameplay is intentionally simple. The main goal is to explore:

- TCP networking with Go.
- Concurrent client connections.
- Goroutines and synchronization.
- Matchmaking.
- Server-authoritative game state.
- Turn validation.
- Connection and disconnection handling.
- A small application protocol.
- Automated tests for game rules.
- Integration with a Unity client.

## Architecture

```text
┌────────────────┐       TCP         ┌──────────────────┐
│ Unity Client A │ ◄───────────────► │                  │
└────────────────┘                   │ Go Heroes Server │
                                     │                  │
┌────────────────┐       TCP         │                  │
│ Unity Client B │ ◄───────────────► │                  │
└────────────────┘                   └──────────────────┘
```

The server owns the canonical match state. Clients send intentions such as `attack`, `defend`, and `heal`; they never decide whether an action is valid or calculate its result.

## MVP flow

```text
Player connects
      │
      ▼
Waiting queue
      │
      │ second player arrives
      ▼
Create match
      │
      ▼
Players take turns
      │
      ▼
Game over
      │
      ▼
Player may enter matchmaking again
```

The MVP has no accounts, database, lobby, room browser, ranking, inventory, or persistent progression.

## Initial protocol

The first version uses **newline-delimited JSON over TCP**.

Client intention:

```json
{"type":"action","action":"attack"}
```

Example server state:

```json
{
  "type": "game_state",
  "yourHp": 80,
  "opponentHp": 60,
  "yourTurn": false
}
```

The exact message schema will evolve during implementation and should be documented here once stabilized.

## Server-authoritative design

The server decides:

- Whose turn it is.
- Whether an action is valid.
- Attack damage.
- Defense effects.
- Healing amount and maximum HP.
- Match state.
- Victory and defeat.

The Unity client is responsible for presentation and input.

## Planned repository structure

The project will start small, not creating layers before it is needed.

```text
go-heroes-server/
├── cmd/
│   └── server/
├── docs
│   └── DEVELOPMENT_PLAN.md
├── internal/
│   ├── game/
│   ├── matchmaking/
│   └── transport/
│       └── tcp/
├── go.mod
└── README.md
```

## Roadmap

### MVP

- TCP listener.
- Multiple concurrent clients.
- JSON message framing.
- Automatic matchmaking.
- Two-player matches.
- Attack, Defend, and Heal.
- Server-enforced turns.
- Win/loss detection.
- Disconnect handling.
- Automated tests for core rules.
- Integration with two Unity clients.

### Future experiments

After the TCP MVP works:

- Lobby and room listing.
- Named/private rooms.
- Real reconnect/session recovery.
- WebSocket transport.
- Browser client.
- gRPC/Protocol Buffers.
- Optional HTTP API.
- PostgreSQL for match history/player statistics.

The active game state should remain in memory until persistence is genuinely useful.

## Development plan

See [`DEVELOPMENT_PLAN.md`](./DEVELOPMENT_PLAN.md).

## Status

**Planning / learning project**

The first milestone is complete when two Unity clients can connect to this server, be automatically matched, play a full server-authoritative match, receive a win/loss result, and queue for another match.

## Related project

[Go Heroes](https://github.com/marcelobruzetti/go-heroes) is the graphical client for this server. The two repositories communicate only through the documented network protocol.