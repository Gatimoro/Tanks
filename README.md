# Tanks

A multiplayer real-time tanks game server written in Go. This is a learning project demonstrating WebSocket communication, concurrent server architecture, and game lobby management.

## Current Project Structure

```
Tanks/
├── main.go           # WebSocket server entry point (port 8080)
├── hub.go            # Central game hub - manages clients & lobbies
├── protocol.go       # Shared message protocol definitions
├── README.md         # This file
└── server/
    ├── client.go     # Individual client WebSocket handler
    └── lobby.go      # Game lobby/room management
```

## Tech Stack

- **Language:** Go (Golang)
- **WebSocket:** github.com/gorilla/websocket
- **Architecture:** Hub-based coordination with goroutines and channels

## Features Implemented

- WebSocket server with connection lifecycle management
- Client registration and session tracking
- Lobby creation, joining, and leaving
- Player ready status tracking
- Real-time message broadcasting
- Thread-safe concurrent operations

## What Needs To Be Done Next

### Priority 1: Critical Fixes (Do These First)

1. **Create `go.mod` file** - The project won't compile without it
   ```bash
   go mod init tankgame
   go mod tidy
   ```

2. **Implement missing Lobby methods** in `server/lobby.go`:
   - `setPlayerReady(client *Client, ready bool)` - Update player ready status
   - `tryStart(client *Client)` - Attempt to start game when all players ready

3. **Fix silent error handling** in `server/client.go:211` - JSON unmarshal errors are currently ignored

### Priority 2: Core Game Features

4. **Implement game state machine** - Transitions between `waiting` → `starting game` → `playing` → `game over`

5. **Add game logic**:
   - Tank movement and positioning
   - Shooting mechanics
   - Collision detection
   - Health/damage system
   - Score tracking

6. **Create game loop** - Server-side tick-based update loop for real-time gameplay

### Priority 3: Client Development

7. **Build a game client** - Options:
   - Web client (HTML5 Canvas + JavaScript)
   - Desktop client (using a Go game library like Ebitengine)

8. **Add client-side prediction** - For smooth gameplay despite network latency

### Priority 4: Production Readiness

9. **Security improvements**:
   - Replace permissive CORS (`CheckOrigin: func(r *http.Request) bool { return true }`)
   - Add client authentication
   - Validate all incoming messages
   - Rate limiting

10. **Add graceful shutdown** - Handle SIGINT/SIGTERM properly

11. **Configuration** - Move hardcoded values to config file or environment variables:
    - Port (currently 8080)
    - Max players per lobby (currently 4)
    - Timeouts and buffer sizes

12. **Logging** - Add structured logging for debugging and monitoring

13. **Testing** - Add unit tests for Hub, Client, and Lobby logic

## Recommended Project Structure (Future)

As the project grows, consider restructuring to:

```
Tanks/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── hub/
│   │   └── hub.go            # Hub logic
│   ├── client/
│   │   └── client.go         # Client handler
│   ├── lobby/
│   │   └── lobby.go          # Lobby management
│   ├── game/
│   │   ├── game.go           # Game state and logic
│   │   ├── tank.go           # Tank entity
│   │   ├── projectile.go     # Bullet/projectile logic
│   │   └── physics.go        # Collision detection
│   └── protocol/
│       └── protocol.go       # Message definitions
├── pkg/
│   └── utils/                # Shared utilities
├── web/                      # Web client (if building one)
│   ├── index.html
│   ├── game.js
│   └── styles.css
├── configs/
│   └── config.yaml           # Server configuration
├── go.mod
├── go.sum
└── README.md
```

## How to Run (After Fixes)

```bash
# Initialize module (first time only)
go mod init tankgame
go mod tidy

# Run the server
go run .

# Server will start on ws://localhost:8080/ws
```

## Message Protocol

### Client → Server
| Type | Description |
|------|-------------|
| `request_lobbies` | Request list of available lobbies |
| `create_lobby` | Create a new game lobby |
| `join_lobby` | Join an existing lobby |
| `leave_lobby` | Leave current lobby |
| `set_ready` | Mark player as ready |
| `start_game` | Initiate game start (host only) |

### Server → Client
| Type | Description |
|------|-------------|
| `lobby_list` | List of available lobbies |
| `lobby_state` | Complete lobby state |
| `player_joined` | Player joined notification |
| `player_left` | Player left notification |
| `error` | Error message |
| `game_starting` | Game is starting notification |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Hub                                  │
│  - Manages all connected clients                            │
│  - Manages all game lobbies                                 │
│  - Routes messages between clients and lobbies              │
└─────────────────┬───────────────────────────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌───────┐    ┌───────┐    ┌───────┐
│Client │    │Client │    │Client │   (Each has 2 goroutines:
│   1   │    │   2   │    │   3   │    readPump + writePump)
└───┬───┘    └───┬───┘    └───┬───┘
    │            │            │
    └────────────┼────────────┘
                 │
                 ▼
           ┌─────────┐
           │  Lobby  │   (Manages game room state,
           │         │    player ready status, etc.)
           └─────────┘
```

## Learning Resources

- [Gorilla WebSocket Examples](https://github.com/gorilla/websocket/tree/master/examples)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Game Loop Patterns](https://gameprogrammingpatterns.com/game-loop.html)
- [Ebitengine (Go Game Library)](https://ebitengine.org/)

## License

This is a learning project. Feel free to use and modify as needed.
