package main

import (
    "errors"
    "sync"

    "tankgame/shared"
)

const maxPlayersPerLobby = 4

type GameState string
const(
	waiting GameState = "waiting"
	countDown GameState = "starting game"
	inGame 	GameState = "playing"
	
)
type Lobby struct {
    id     string
    name   string
    hostID string // Player who created it (can start game)
    state  GameState// True once game starts

    players map[string]*Client // playerID -> Client
    ready   map[string]bool    // playerID -> ready status
    mu      sync.RWMutex       // Protects players and ready maps

    hub *Hub // Reference back to hub (to remove self when empty)
}

func NewLobby(id, name string, host *Client, hub *Hub) *Lobby {
    l := &Lobby{
        id:      id,
        name:    name,
        hostID:  host.id,
        players: make(map[string]*Client),
        ready:   make(map[string]bool),
        hub:     hub,
    }
    // Host is automatically in the lobby
    l.players[host.id] = host
    l.ready[host.id] = false
    return l
}
func (l *Lobby) addPlayer(c *Client) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    if l.inGame {
        return errors.New("game already in progress")
    }

    if len(l.players) >= maxPlayersPerLobby {
        return errors.New("lobby is full")
    }

    if _, exists := l.players[c.id]; exists {
        return errors.New("already in this lobby")
    }

    // Add the player
    l.players[c.id] = c
    l.ready[c.id] = false

    // Tell everyone already in the lobby that someone joined
    l.broadcastUnlocked(shared.Envelope{
        Type: shared.MsgPlayerJoined,
        Payload: shared.PlayerJoinedPayload{
            Player: shared.PlayerInfo{
                ID:     c.id,
                Name:   c.name,
                Ready:  false,
                IsHost: false,
            },
        },
    })

    // Send lobby state to the new player
    l.sendStateToClientUnlocked(c)

    return nil
}
func (l *Lobby) removePlayer(c *Client) {
    l.mu.Lock()
    defer l.mu.Unlock()

    if _, exists := l.players[c.id]; !exists {
        return // Not in this lobby
    }

    delete(l.players, c.id)
    delete(l.ready, c.id)

    // If lobby is now empty, remove it from hub
    if len(l.players) == 0 {
        l.hub.removeLobby(l.id)
        return
    }

    // If the host left, pick a new host
    if l.hostID == c.id {
        for id := range l.players {
            l.hostID = id // Just pick the first one
            break
        }
    }

    // Tell remaining players someone left
    l.broadcastUnlocked(shared.Envelope{
        Type:    shared.MsgPlayerLeft,
        Payload: shared.PlayerLeftPayload{PlayerID: c.id},
    })

    // Send updated state (in case host changed)
    l.sendStateToAllUnlocked()
}

// Info returns a summary for the lobby list (shown in menu)
func (l *Lobby) Info() shared.LobbyInfo {
    l.mu.RLock()
    defer l.mu.RUnlock()

    return shared.LobbyInfo{
        ID:          l.id,
        Name:        l.name,
        PlayerCount: len(l.players),
        MaxPlayers:  maxPlayersPerLobby,
        InGame:      l.inGame,
    }
}

// sendStateToAll sends full lobby state to every player
func (l *Lobby) sendStateToAll() {
    l.mu.RLock()
    defer l.mu.RUnlock()
    l.sendStateToAllUnlocked()
}

func (l *Lobby) sendStateToAllUnlocked() {
    for _, client := range l.players {
        l.sendStateToClientUnlocked(client)
    }
}

func (l *Lobby) sendStateToClientUnlocked(c *Client) {
    players := make([]shared.PlayerInfo, 0, len(l.players))
    for id, client := range l.players {
        players = append(players, shared.PlayerInfo{
            ID:     id,
            Name:   client.name,
            Ready:  l.ready[id],
            IsHost: id == l.hostID,
        })
    }

    c.sendEnvelope(shared.Envelope{
        Type: shared.MsgLobbyState,
        Payload: shared.LobbyStatePayload{
            LobbyID:    l.id,
            LobbyName:  l.name,
            Players:    players,
            YouAreHost: c.id == l.hostID,
        },
    })
}

func (l *Lobby) broadcastUnlocked(env shared.Envelope) {
    for _, client := range l.players {
        client.sendEnvelope(env)
    }
}
