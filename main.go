package main

import (
    "crypto/rand"
    "encoding/hex"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins (dev only!)
}

// generateID creates a random hex string for IDs
func generateID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// serveWs handles a new WebSocket connection
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade error:", err)
        return
    }

    // Create client, register with hub
    client := NewClient(generateID(), hub, conn)
    hub.register <- client

    log.Printf("client connected: %s", client.id)

    // Start the pumps (each in their own goroutine)
    go client.writePump()
    go client.readPump() // This one blocks until disconnect
}

func main() {
    hub := NewHub()
    go hub.Run() // Start hub in background

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(hub, w, r)
    })

    addr := ":8080"
    log.Printf("server starting on %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
