package gemini

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// CachedImageCookies holds pre-fetched cookies for image/video downloads
var CachedImageCookies string

// OnCookiesUpdated is a callback triggered when new cookies are synced
var OnCookiesUpdated func()

var (
	activeClients   []*websocket.Conn
	activeClientsMu sync.Mutex
)

// BroadcastCookieRefresh sends a trigger_sync message to all connected Chrome Extension clients
func BroadcastCookieRefresh() {
	activeClientsMu.Lock()
	defer activeClientsMu.Unlock()
	log.Printf("📢 Broadcasting trigger_sync to %d connected Chrome Extensions...", len(activeClients))

	payload := map[string]string{"type": "trigger_sync"}

	// We iterate backwards so we can safely remove disconnected clients
	for i := len(activeClients) - 1; i >= 0; i-- {
		conn := activeClients[i]
		err := conn.WriteJSON(payload)
		if err != nil {
			log.Printf("⚠️ Failed to write to extension client: %v. Removing client.", err)
			conn.Close()
			activeClients = append(activeClients[:i], activeClients[i+1:]...)
		}
	}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow Chrome Extension context
	},
}

type ExtensionMessage struct {
	Type    string         `json:"type"`
	Cookies []CookieObject `json:"cookies"`
}

// StartCookieWebSocketServer starts a local WebSocket server to receive cookies from the extension
func StartCookieWebSocketServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("❌ WS Upgrade failed: %v", err)
			return
		}

		activeClientsMu.Lock()
		activeClients = append(activeClients, conn)
		activeClientsMu.Unlock()

		defer func() {
			conn.Close()
			activeClientsMu.Lock()
			for i, c := range activeClients {
				if c == conn {
					activeClients = append(activeClients[:i], activeClients[i+1:]...)
					break
				}
			}
			activeClientsMu.Unlock()
			log.Println("🔌 Chrome Extension disconnected")
		}()

		log.Println("🔌 Chrome Extension connected to cookie bridge")

		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println("🔌 Chrome Extension disconnected")
				break
			}

			var msg ExtensionMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Printf("❌ Failed to decode extension message: %v", err)
				continue
			}

			if msg.Type == "ping" {
				conn.WriteJSON(map[string]string{"type": "pong"})
				continue
			}

			if msg.Type == "cookies_payload" && len(msg.Cookies) > 0 {
				log.Printf("🍪 Received %d cookies from Chrome Extension", len(msg.Cookies))

				// Write to cookies.json
				data, err := json.MarshalIndent(msg.Cookies, "", "  ")
				if err != nil {
					log.Printf("❌ Failed to marshal cookies: %v", err)
					continue
				}

				if err := os.WriteFile("cookies.json", data, 0644); err != nil {
					log.Printf("❌ Failed to save cookies.json: %v", err)
					continue
				}

				// Refresh CachedImageCookies
				var parts []string
				for _, ck := range msg.Cookies {
					parts = append(parts, fmt.Sprintf("%s=%s", ck.Name, ck.Value))
				}
				CachedImageCookies = strings.Join(parts, "; ")
				log.Printf("🍪 Cached %d cookies for image/video downloading", len(msg.Cookies))

				// Trigger callback to reload active sessions
				if OnCookiesUpdated != nil {
					OnCookiesUpdated()
				}
			}
		}
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("📡 Cookie WebSocket Server listening on %s", addr)
	
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Cookie WebSocket Server failed: %v", err)
	}
}

