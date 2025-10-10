package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// CHANGE THESE
const (
	port       = ":8080"
	allowedURL = "https://YOURWEBSITE.com"
)

var (
	// current stored scratchpad content
	content     = ""
	contentLock sync.RWMutex

	// connected clients
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex

	// WebSocket upgrader
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == allowedURL
		},
	}
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)

	log.Printf("Server starting on %s (allowing origin: %s)\n", port, allowedURL)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// main page is just index.html
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

// the /ws endpoint works over WebSocket API
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// HTTP -> WebSocket upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// register new client
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// send current content to new client
	contentLock.RLock()
	currentContent := content
	contentLock.RUnlock()
	
	if err := conn.WriteJSON(Message{Type: "update", Content: currentContent}); err != nil {
		log.Println("Write error:", err)
		return
	}

	// kill client on disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	// wait and listen for client messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == "update" {
			contentLock.Lock()
			content = msg.Content
			contentLock.Unlock()
			broadcast(msg)
		}
	}
}

func broadcast(msg Message) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("Broadcast error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}