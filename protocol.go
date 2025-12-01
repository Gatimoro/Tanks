package shared

// MessageType identifies what kind of message this is
type MessageType string

const (
    // Client -> Server 
    MsgRequestLobbies MessageType = "request_lobbies"
    MsgCreateLobby    MessageType = "create_lobby"
    MsgJoinLobby      MessageType = "join_lobby"
    MsgLeaveLobby     MessageType = "leave_lobby"
    MsgSetReady       MessageType = "set_ready"
    MsgStartGame      MessageType = "start_game"

    // Server -> Client 
    MsgLobbyList    MessageType = "lobby_list"
    MsgLobbyState   MessageType = "lobby_state"
    MsgPlayerJoined MessageType = "player_joined"
    MsgPlayerLeft   MessageType = "player_left"
    MsgError        MessageType = "error"
    MsgGameStarting MessageType = "game_starting"
)

// Envelope wraps all messages - this is what actually goes over the wire
type Envelope struct {
    Type    MessageType `json:"type"`
    Payload any         `json:"payload,omitempty"`
}
// === Payloads for Client -> Server ===

type CreateLobbyPayload struct {
    Name string `json:"name"`
}

type JoinLobbyPayload struct {
    LobbyID string `json:"lobby_id"`
}

type SetReadyPayload struct {
    Ready bool `json:"ready"`
}

// === Payloads for Server -> Client ===

type LobbyInfo struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    PlayerCount int    `json:"player_count"`
    MaxPlayers  int    `json:"max_players"`
    InGame      bool   `json:"in_game"`
}

type LobbyListPayload struct {
    Lobbies []LobbyInfo `json:"lobbies"`
}

type PlayerInfo struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Ready  bool   `json:"ready"`
    IsHost bool   `json:"is_host"`
}

type LobbyStatePayload struct {
    LobbyID    string       `json:"lobby_id"`
    LobbyName  string       `json:"lobby_name"`
    Players    []PlayerInfo `json:"players"`
    YouAreHost bool         `json:"you_are_host"`
}

type PlayerJoinedPayload struct {
    Player PlayerInfo `json:"player"`
}

type PlayerLeftPayload struct {
    PlayerID string `json:"player_id"`
}

type ErrorPayload struct {
    Message string `json:"message"`
}

type GameStartingPayload struct {
    YourTankID string `json:"your_tank_id"`
}
