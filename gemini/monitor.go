package gemini

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

// GetActiveWorkerCount returns number of connected Chrome Extension workers
func GetActiveWorkerCount() int {
	activeClientsMu.Lock()
	defer activeClientsMu.Unlock()
	return len(activeClients)
}

func getAccountIDFromCookies(cookies []CookieObject) string {
	for _, c := range cookies {
		if c.Name == "__Secure-1PSID" || c.Name == "SID" || c.Name == "HSID" {
			cleanVal := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(c.Value, "")
			if len(cleanVal) > 10 {
				return cleanVal[:10]
			}
			return cleanVal
		}
	}
	return "primary"
}

// GetAvailableAccountCookieFiles returns paths of all active account cookie files
func GetAvailableAccountCookieFiles() []string {
	files, err := filepath.Glob(filepath.Join("cookies", "account_*.json"))
	if err != nil || len(files) == 0 {
		defaultPath := filepath.Join("cookies", "cookies.json")
		if _, err := os.Stat(defaultPath); err == nil {
			return []string{defaultPath}
		}
		return nil
	}
	return files
}

// GetActiveAccountCount returns the number of distinct account profiles saved
func GetActiveAccountCount() int {
	return len(GetAvailableAccountCookieFiles())
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
		count := len(activeClients)
		activeClientsMu.Unlock()

		log.Printf("🔌 Chrome Extension connected to cookie bridge (Active Workers: %d)", count)

		defer func() {
			conn.Close()
			activeClientsMu.Lock()
			for i, c := range activeClients {
				if c == conn {
					activeClients = append(activeClients[:i], activeClients[i+1:]...)
					break
				}
			}
			remCount := len(activeClients)
			activeClientsMu.Unlock()
			log.Printf("🔌 Chrome Extension disconnected (Active Workers: %d)", remCount)
		}()

		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
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
				accountID := getAccountIDFromCookies(msg.Cookies)
				log.Printf("🍪 Received %d cookies for Account [%s] from Chrome Extension", len(msg.Cookies), accountID)

				data, err := json.MarshalIndent(msg.Cookies, "", "  ")
				if err != nil {
					log.Printf("❌ Failed to marshal cookies: %v", err)
					continue
				}

				os.MkdirAll("cookies", 0755)

				// Save account-specific cookie file for multi-account pool
				accountFilePath := filepath.Join("cookies", fmt.Sprintf("account_%s.json", accountID))
				if err := os.WriteFile(accountFilePath, data, 0644); err != nil {
					log.Printf("❌ Failed to save %s: %v", accountFilePath, err)
				} else {
					log.Printf("💾 Stored account profile: %s", accountFilePath)
				}

				// Also write to default cookies/cookies.json
				defaultCookiePath := filepath.Join("cookies", "cookies.json")
				_ = os.WriteFile(defaultCookiePath, data, 0644)

				// Refresh CachedImageCookies
				var parts []string
				for _, ck := range msg.Cookies {
					parts = append(parts, fmt.Sprintf("%s=%s", ck.Name, ck.Value))
				}
				CachedImageCookies = strings.Join(parts, "; ")

				// Trigger callback to reload active sessions
				if OnCookiesUpdated != nil {
					OnCookiesUpdated()
				}
			}
		}
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("📡 Cookie WebSocket Server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Cookie WebSocket Server failed: %v", err)
	}
}
