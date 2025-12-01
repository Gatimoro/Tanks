package main

import (
    "encoding/json"
    "log"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "tankgame/shared"
)

const (
    writeWait      = 10 * time.Second    // Max time to write a message
    pongWait       = 60 * time.Second    // Max time to wait for pong
    pingPeriod     = (pongWait * 9) / 10 // How often to ping
    maxMessageSize = 4096
)

// Client represents a single connected player
type Client struct {
    id   string          // Unique identifier (generated on connect)
    name string          // Display name like "Player-abc123"
    hub  *Hub            // Reference to the central hub (we'll build this)
    conn *websocket.Conn // The actual WebSocket connection

    send chan []byte     // Buffered channel for outgoing messages

    // What lobby this client is in (nil if in menu)
    lobby   *Lobby
    lobbyMu sync.RWMutex // Protects lobby field from race conditions
}

func NewClient(id string, hub *Hub, conn *websocket.Conn) *Client {
    return &Client{
        id:   id,
        name: "Player-" + id[:6], // Short readable name from ID
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 64), // Buffer 64 messages
    }
}
// readPump reads messages from the WebSocket and processes them
// Runs in its own goroutine - one per client
func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c  // Tell hub we're disconnecting
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    // This loop runs forever until connection breaks
    for {
        _, message, err := c.conn.ReadMessage()  // BLOCKS here waiting for data
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("client %s read error: %v", c.id, err)
            }
            break  //and cleanup from defer
        }

        c.handleMessage(message)  // Process the message TODO
    }
}

// sends messages from the send channel to the WebSocket
// Runs in its own goroutine - one per client
func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // we're being kicked
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                log.Printf("client %s write error: %v", c.id, err)
                return
            }

        case <-ticker.C:
            // periodic ping to keep connection alive
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
// ------------------------------------------------------------------------
//			HANDLERS
// ------------------------------------------------------------------------

//processes an incoming message from this client
func (c *Client) handleMessage(data []byte) {
    var env shared.Envelope
    if err := json.Unmarshal(data, &env); err != nil {
        log.Printf("client %s invalid json: %v", c.id, err)
        c.sendError("invalid message format")
        return
    }

    log.Printf("client %s sent: %s", c.id, env.Type)

    // Route based on message type
    switch env.Type {
    case shared.MsgRequestLobbies:
        c.hub.sendLobbyList(c)

    case shared.MsgCreateLobby:
        c.handleCreateLobby(env.Payload)

    case shared.MsgJoinLobby:
        c.handleJoinLobby(env.Payload)

    case shared.MsgLeaveLobby:
        c.handleLeaveLobby()

    case shared.MsgSetReady:
        c.handleSetReady(env.Payload)

    case shared.MsgStartGame:
        c.handleStartGame()

    default:
        c.sendError("unknown message type: " + string(env.Type))
    }
}

func (c *Client) handleCreateLobby(payload any) {
    // payload comes as map[string]any from JSON unmarshaling
    // We need to convert it to our typed struct
    data, _ := json.Marshal(payload)
    var p shared.CreateLobbyPayload
    if err := json.Unmarshal(data, &p); err != nil || p.Name == "" {
        c.sendError("invalid lobby name")
        return
    }

    // Leave current lobby if in one
    c.handleLeaveLobby()

    // Create the lobby (we become host)
    lobby := c.hub.createLobby(p.Name, c)
    c.setLobby(lobby)

    // Send lobby state to us (and anyone else, but it's just us)
    lobby.sendStateToAll()
}
func (c *Client) handleJoinLobby(payload any) {
    data, _ := json.Marshal(payload)
    var p shared.JoinLobbyPayload
    if err := json.Unmarshal(data, &p); err != nil || p.LobbyID == "" {
        c.sendError("invalid lobby id")
        return
    }

    // Leave current lobby first
    c.handleLeaveLobby()

    // Find the lobby
    lobby := c.hub.getLobby(p.LobbyID)
    if lobby == nil {
        c.sendError("lobby not found")
        return
    }

    // Try to join (might fail if full or in-game)
    if err := lobby.addPlayer(c); err != nil {
        c.sendError(err.Error())
        return
    }

    // Success - update our lobby reference
    c.setLobby(lobby)
    // Note: addPlayer already sent us the lobby state
}
func (c *Client) handleLeaveLobby() {
	c.lobbyMu.Lock()
	lobby := c.lobby
	c.lobby = nil
	c.lobbyMu.Unlock()

	if lobby != nil {
		lobby.removePlayer(c)
	}
}

func (c *Client) handleSetReady(payload any) {
    data, err := json.Marshal(payload)
    if err != nil {
    log.Printf("marshal error: %v", err)  // Log it at least
    	return
    }
    var p shared.SetReadyPayload
    json.Unmarshal(data, &p)

    c.lobbyMu.RLock()
    lobby := c.lobby
    c.lobbyMu.RUnlock()

    if lobby != nil {
        lobby.setPlayerReady(c, p.Ready)
    }
}

func (c *Client) handleStartGame() {
    c.lobbyMu.RLock()
    lobby := c.lobby
    c.lobbyMu.RUnlock()

    if lobby != nil {
        lobby.tryStart(c)
    }
}

func (c *Client) setLobby(l *Lobby) {
    c.lobbyMu.Lock()
    c.lobby = l
    c.lobbyMu.Unlock()
}

// marshals and queues a message to be sent
func (c *Client) sendEnvelope(env shared.Envelope) {
    data, err := json.Marshal(env)
    if err != nil {
        log.Printf("marshal error: %v", err)
        return
    }

    select {
    case c.send <- data:
        // Queued 
    default:
// Channel full 
        log.Printf("client %s send buffer full, dropping", c.id)
    }
}

func (c *Client) sendError(msg string) {
    c.sendEnvelope(shared.Envelope{
        Type:    shared.MsgError,
        Payload: shared.ErrorPayload{Message: msg},
    })
}
