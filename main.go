package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"goapi/gemini"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatCompletionRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type OpenAIChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type OpenAIChatCompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}

type OpenAIDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type OpenAIStreamChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type OpenAIChatCompletionChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
}

var (
	userSessions sync.Map
)

const CookiesFile = "cookies.json"

func downloadAndClean(client *gemini.GeminiClient, urlStr, filename, fileType string) error {
	tempDir := "./output/.temp"
	os.MkdirAll(tempDir, 0755)
	os.MkdirAll("./output", 0755)

	tempPath := filepath.Join(tempDir, filename)
	outputPath := filepath.Join("./output", filename)

	// Step 1: Download to temp (which automatically cleans the watermark natively in Go!)
	if err := client.DownloadFile(urlStr, tempPath); err != nil {
		os.RemoveAll(tempDir)
		return err
	}

	// Step 2: Move from temp to output folder
	var moveErr error
	if err := os.Rename(tempPath, outputPath); err != nil {
		// Fallback: Copy + Delete if rename fails across different mount points
		input, err := os.ReadFile(tempPath)
		if err != nil {
			moveErr = err
		} else if err := os.WriteFile(outputPath, input, 0644); err != nil {
			moveErr = err
		}
	}

	// Clean up entire temp directory
	os.RemoveAll(tempDir)

	if moveErr != nil {
		return moveErr
	}

	log.Printf("✨ Cleaned file moved to output: %s", outputPath)
	return nil
}

func getOrCreateClient(sessionID string) (*gemini.GeminiClient, error) {
	if client, ok := userSessions.Load(sessionID); ok {
		return client.(*gemini.GeminiClient), nil
	}

	if _, err := os.Stat(CookiesFile); os.IsNotExist(err) {
		log.Println("⚠️ cookies.json not found. Requesting Chrome Extension to proactively sync cookies...")
		gemini.BroadcastCookieRefresh()

		log.Println("⏳ Sleeping 5 seconds waiting for extension to sync cookies...")
		time.Sleep(5 * time.Second)

		if _, err := os.Stat(CookiesFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("cookies.json not found. Please ensure Chrome Extension is connected and syncs cookies first.")
		}
	}

	client, err := gemini.NewClient(CookiesFile)
	if err != nil {
		return nil, err
	}

	userSessions.Store(sessionID, client)
	log.Printf("New session created for: %s", sessionID)
	return client, nil
}

func startWebSocketBridge() {
	// Initialize callback to reload sessions when extension updates cookies
	gemini.OnCookiesUpdated = func() {
		log.Println("♻️  Extension pushed new cookies. Reloading active sessions...")
		userSessions.Range(func(key, value interface{}) bool {
			client := value.(*gemini.GeminiClient)
			sessionID := key.(string)

			if err := client.ReloadSession(); err != nil {
				log.Printf("❌ Failed to reload session for %s: %v", sessionID, err)
			} else {
				log.Printf("✅ Session %s refreshed", sessionID)
			}
			return true
		})
	}

	wsPortStr := os.Getenv("WS_PORT")
	if wsPortStr == "" {
		wsPortStr = "9226" // default port matching the extension target
	}
	wsPort, err := strconv.Atoi(wsPortStr)
	if err != nil {
		wsPort = 9226
	}

	// Start WebSocket Server
	go gemini.StartCookieWebSocketServer(wsPort)
}

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("⚠️  No config.env file found")
	}

	startWebSocketBridge()


	app := fiber.New(fiber.Config{
		AppName:   "Gemini Go API",
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// Add recovery to prevent crashes and logger for debugging
	app.Use(logger.New())
	app.Use(func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ RECOVERED from panic: %v", r)
				c.Status(500).JSON(fiber.Map{"error": "Internal Server Error - Recovered"})
			}
		}()
		return c.Next()
	})

	// Unified chat endpoint - handles text, images, video, everything
	app.Post("/chat", func(c fiber.Ctx) error {
		var prompt, userID string
		var newChat, stream bool
		var images []gemini.ImageInput

		contentType := string(c.Request().Header.ContentType())

		if strings.Contains(contentType, "multipart/form-data") {
			// Multipart: image upload mode
			prompt = c.FormValue("prompt")
			userID = c.FormValue("user_id")
			newChat = c.FormValue("new_chat") == "true"
			stream = c.FormValue("stream") == "true"

			form, err := c.MultipartForm()
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid multipart form"})
			}

			var fileHeaders []*multipart.FileHeader
			if files, ok := form.File["image"]; ok {
				fileHeaders = append(fileHeaders, files...)
			}
			if files, ok := form.File["images"]; ok {
				fileHeaders = append(fileHeaders, files...)
			}

			if len(fileHeaders) > 10 {
				return c.Status(400).JSON(fiber.Map{"error": "Maximum 10 images allowed per request"})
			}

			for _, fh := range fileHeaders {
				file, err := fh.Open()
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to open image file"})
				}
				data, err := io.ReadAll(file)
				file.Close()
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to read image data"})
				}
				mime := fh.Header.Get("Content-Type")
				if mime == "" {
					mime = "image/jpeg"
				}
				images = append(images, gemini.ImageInput{Data: data, Filename: fh.Filename, MimeType: mime})
			}
		} else {
			// JSON: text-only mode
			var req gemini.ChatRequest
			if err := c.Bind().JSON(&req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
			}
			prompt = req.Prompt
			userID = req.UserID
			newChat = req.NewChat
			stream = req.Stream
		}

		sessionID := userID
		if sessionID == "" {
			sessionID = c.IP()
		}

		if newChat {
			userSessions.Delete(sessionID)
		}

		client, err := getOrCreateClient(sessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		var resp *gemini.GeminiResponse

		if stream && len(images) == 0 {
			// SSE Streaming mode (text-only)
			c.Set("Content-Type", "text/event-stream")
			c.Set("Cache-Control", "no-cache")
			c.Set("Connection", "keep-alive")
			c.Set("X-Accel-Buffering", "no")

			return c.SendStreamWriter(func(w *bufio.Writer) {
				var streamResp *gemini.GeminiResponse
				var streamErr error

				streamResp, streamErr = client.AskStream(prompt, func(chunk string) {
					data, _ := json.Marshal(fiber.Map{"text": chunk})
					fmt.Fprintf(w, "data: %s\n\n", data)
					w.Flush()
				})

				if streamErr != nil {
					data, _ := json.Marshal(fiber.Map{"error": streamErr.Error()})
					fmt.Fprintf(w, "data: %s\n\n", data)
					w.Flush()
					return
				}

				// Download images locally
				if len(streamResp.Images) > 0 {
					for i, imgURL := range streamResp.Images {
						filename := fmt.Sprintf("img_%s_%d.png", streamResp.ResponseID, i)
						if streamResp.ResponseID == "" {
							filename = fmt.Sprintf("img_%d_%d.png", time.Now().Unix(), i)
						}
						if err := downloadAndClean(client, imgURL, filename, "image"); err == nil {
							streamResp.Images[i] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
						}
					}
				}

				// Download videos locally
				if len(streamResp.Videos) > 0 {
					vidURL := streamResp.Videos[0]
					if vidURL != "" && len(vidURL) > 50 {
						filename := fmt.Sprintf("vid_%s_0.mp4", streamResp.ResponseID)
						if streamResp.ResponseID == "" {
							filename = fmt.Sprintf("vid_%d_0.mp4", time.Now().Unix())
						}
						if err := downloadAndClean(client, vidURL, filename, "video"); err == nil {
							streamResp.Videos[0] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
						}
					}
				}

				// Final event with full response
				final, _ := json.Marshal(streamResp)
				fmt.Fprintf(w, "data: %s\n\n", final)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.Flush()
			})
		}

		if len(images) > 0 {
			// Image mode: upload + chat (auto video polling inside)
			log.Printf("📸 Image request: %d images, prompt: %s", len(images), prompt)
			resp, err = client.AskWithImages(prompt, images)
		} else {
			// Text-only mode
			resp, err = client.Ask(prompt)
		}

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				userSessions.Delete(sessionID)
				log.Printf("🔄 Session %s auto-reset due to timeout", sessionID)
				return c.Status(504).JSON(fiber.Map{"error": "Request timed out. Session reset."})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Download images locally
		if len(resp.Images) > 0 {
			for i, imgURL := range resp.Images {
				filename := fmt.Sprintf("img_%s_%d.png", resp.ResponseID, i)
				if resp.ResponseID == "" {
					filename = fmt.Sprintf("img_%d_%d.png", time.Now().Unix(), i)
				}
				if err := downloadAndClean(client, imgURL, filename, "image"); err == nil {
					resp.Images[i] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
				} else {
					log.Printf("❌ Failed to download image: %v", err)
				}
			}
		}

		// Download videos locally
		if len(resp.Videos) > 0 {
			vidURL := resp.Videos[0]
			if vidURL != "" && len(vidURL) > 50 {
				filename := fmt.Sprintf("vid_%s_0.mp4", resp.ResponseID)
				if resp.ResponseID == "" {
					filename = fmt.Sprintf("vid_%d_0.mp4", time.Now().Unix())
				}
				if err := downloadAndClean(client, vidURL, filename, "video"); err == nil {
					resp.Videos[0] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
				}
			}
		}

		return c.JSON(resp)
	})

	// OpenAI-compatible endpoint for drop-in client integrations
	app.Post("/v1/chat/completions", func(c fiber.Ctx) error {
		var req OpenAIChatCompletionRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if len(req.Messages) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Messages array is empty"})
		}

		// Extract prompt from messages (last user message)
		var prompt string
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				prompt = req.Messages[i].Content
				break
			}
		}
		if prompt == "" {
			prompt = req.Messages[len(req.Messages)-1].Content
		}

		sessionID := c.IP()

		client, err := getOrCreateClient(sessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		model := req.Model
		if model == "" {
			model = "gemini-2.5-flash"
		}

		if req.Stream {
			c.Set("Content-Type", "text/event-stream")
			c.Set("Cache-Control", "no-cache")
			c.Set("Connection", "keep-alive")
			c.Set("X-Accel-Buffering", "no")

			return c.SendStreamWriter(func(w *bufio.Writer) {
				chunkID := "chatcmpl-" + uuid.NewString()[:12]
				createdTime := time.Now().Unix()

				_, streamErr := client.AskStream(prompt, func(chunk string) {
					chunkResp := OpenAIChatCompletionChunk{
						ID:      chunkID,
						Object:  "chat.completion.chunk",
						Created: createdTime,
						Model:   model,
						Choices: []OpenAIStreamChoice{
							{
								Index: 0,
								Delta: OpenAIDelta{
									Content: chunk,
								},
							},
						},
					}
					data, _ := json.Marshal(chunkResp)
					fmt.Fprintf(w, "data: %s\n\n", data)
					w.Flush()
				})

				if streamErr != nil {
					log.Printf("❌ Streaming error: %v", streamErr)
					return
				}

				// Send final stop chunk
				finalResp := OpenAIChatCompletionChunk{
					ID:      chunkID,
					Object:  "chat.completion.chunk",
					Created: createdTime,
					Model:   model,
					Choices: []OpenAIStreamChoice{
						{
							Index:        0,
							FinishReason: "stop",
						},
					},
				}
				data, _ := json.Marshal(finalResp)
				fmt.Fprintf(w, "data: %s\n\n", data)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.Flush()
			})
		}

		resp, err := client.Ask(prompt)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				userSessions.Delete(sessionID)
				return c.Status(504).JSON(fiber.Map{"error": "Request timed out. Session reset."})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		openAIResp := OpenAIChatCompletionResponse{
			ID:      "chatcmpl-" + uuid.NewString()[:12],
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []OpenAIChoice{
				{
					Index: 0,
					Message: OpenAIChatMessage{
						Role:    "assistant",
						Content: resp.Text,
					},
					FinishReason: "stop",
				},
			},
		}

		return c.JSON(openAIResp)
	})

	// Music generation endpoint (separate because it uses tool="music_gen")
	app.Post("/music", func(c fiber.Ctx) error {
		var req gemini.ChatRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		sessionID := req.UserID
		if sessionID == "" {
			sessionID = c.IP()
		}

		if req.NewChat {
			userSessions.Delete(sessionID)
		}

		client, err := getOrCreateClient(sessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		log.Printf("🎵 Music request from %s: %s", sessionID, req.Prompt)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		resp, err := client.AskWithTool(req.Prompt, "music_gen")
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				userSessions.Delete(sessionID)
				return c.Status(504).JSON(fiber.Map{"error": "Music generation timed out."})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if len(resp.Music) > 0 {
			var filteredMusic []gemini.MusicTrack
			for i, track := range resp.Music {
				if track.DownloadURL == "" || len(track.DownloadURL) < 50 {
					continue
				}
				ext := ".mp3"
				if strings.Contains(track.DownloadURL, ".mp4") {
					if i > 0 {
						continue
					}
					ext = ".mp4"
				}
				filename := fmt.Sprintf("music_%s_%d%s", resp.ResponseID, i, ext)
				if resp.ResponseID == "" {
					filename = fmt.Sprintf("music_%d_%d%s", time.Now().Unix(), i, ext)
				}
				if err := downloadAndClean(client, track.DownloadURL, filename, "music"); err == nil {
					track.LocalPath = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
					log.Printf("🎵 Music downloaded: %s", track.Title)
					filteredMusic = append(filteredMusic, track)
					break
				} else {
					log.Printf("❌ Failed to download music track %d: %v", i, err)
				}
			}
			resp.Music = filteredMusic
		}

		return c.JSON(resp)
	})

	app.Post("/reset", func(c fiber.Ctx) error {
		type ResetRequest struct {
			UserID string `json:"user_id"`
		}
		var req ResetRequest
		c.Bind().JSON(&req)

		sessionID := req.UserID
		if sessionID == "" {
			sessionID = c.IP()
		}

		if _, ok := userSessions.Load(sessionID); ok {
			userSessions.Delete(sessionID)
			return c.JSON(fiber.Map{"status": "success", "message": "Session reset"})
		}

		return c.JSON(fiber.Map{"status": "error", "message": "No active session"})
	})

	app.Get("/status", func(c fiber.Ctx) error {
		sessionID := c.Query("user_id")
		if sessionID == "" {
			sessionID = c.IP()
		}

		if clientRaw, ok := userSessions.Load(sessionID); ok {
			client := clientRaw.(*gemini.GeminiClient)
			return c.JSON(fiber.Map{
				"session_id":      sessionID,
				"initialized":     client.IsInitialized,
				"conversation_id": client.ConversationID,
				"expired":         false,
			})
		}
		return c.JSON(fiber.Map{"session_id": sessionID, "active": false})
	})

	app.Get("/output/*", func(c fiber.Ctx) error {
		return c.SendFile("./output/" + c.Params("*"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Fatal(app.Listen(":" + port))
}
