package main

import (
	"sync"
	"tankgame/shared"
)

// Active clients and labels
type Hub struct {
    // id -> Client
    clients   map[string]*Client
    clientsMu sync.RWMutex

    // id -> Lobby
    lobbies   map[string]*Lobby
    lobbiesMu sync.RWMutex

    // client registration 
    register   chan *Client
    unregister chan *Client
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]*Client),
        lobbies:    make(map[string]*Lobby),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}
// Run starts the hub's main loop 
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clientsMu.Lock()
            h.clients[client.id] = client
            h.clientsMu.Unlock()
            log.Printf("client registered: %s (total: %d)", client.id, len(h.clients))

        case client := <-h.unregister:
            h.clientsMu.Lock()
            if _, ok := h.clients[client.id]; ok {
                // Clean up: leave lobby if in one
                client.handleLeaveLobby()
                // Close send channel (signals writePump to exit)
                close(client.send)
                delete(h.clients, client.id)
                log.Printf("client unregistered: %s (total: %d)", client.id, len(h.clients))
            }
            h.clientsMu.Unlock()
        }
    }
}
// new lobby with given host
func (h *Hub) createLobby(name string, host *Client) *Lobby {
    h.lobbiesMu.Lock()
    defer h.lobbiesMu.Unlock()

    id := generateID()  // We'll write this helper
    lobby := NewLobby(id, name, host, h)
    h.lobbies[id] = lobby

    log.Printf("lobby created: %s (%s)", name, id)
    return lobby
}

// retrieves a lobby by ID, or nil if not found
func (h *Hub) getLobby(id string) *Lobby {
    h.lobbiesMu.RLock()
    defer h.lobbiesMu.RUnlock()
    return h.lobbies[id]
}

// deletes a lobby (called when it becomes empty)
func (h *Hub) removeLobby(id string) {
    h.lobbiesMu.Lock()
    defer h.lobbiesMu.Unlock()
    delete(h.lobbies, id)
    log.Printf("lobby removed: %s", id)
}

// sendLobbyList sends the current list of lobbies to a client
func (h *Hub) sendLobbyList(c *Client) {
    h.lobbiesMu.RLock()
    defer h.lobbiesMu.RUnlock()

    lobbies := make([]shared.LobbyInfo, 0, len(h.lobbies))
    for _, lobby := range h.lobbies {
        lobbies = append(lobbies, lobby.Info())
    }

    c.sendEnvelope(shared.Envelope{
        Type:    shared.MsgLobbyList,
        Payload: shared.LobbyListPayload{Lobbies: lobbies},
    })
}
