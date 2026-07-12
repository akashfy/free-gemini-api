package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── JSON-RPC 2.0 Protocol Types ──────────────────────────────────────────────

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ─── MCP Tool Types ───────────────────────────────────────────────────────────

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type PromptArg struct {
	Prompt         string `json:"prompt"`
	RefImagePath   string `json:"ref_image_path,omitempty"`
	RefVideoPath   string `json:"ref_video_path,omitempty"`
	StartImagePath string `json:"start_image_path,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type CallToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ─── Configuration ────────────────────────────────────────────────────────────

const (
	serverName    = "free-gemini-mcp"
	serverVersion = "2.1.0"

	maxRetries    = 3
	baseRetryWait = 2 * time.Second
	httpTimeout   = 5 * time.Minute

	// 10MB scanner buffer for large JSON-RPC messages (base64 images etc.)
	maxScannerBuffer = 10 * 1024 * 1024
)

var apiURL = getAPIURL()

func getAPIURL() string {
	if u := os.Getenv("GEMINI_API_URL"); u != "" {
		return u
	}
	return "http://localhost:8002"
}

// ─── Main Entry Point ─────────────────────────────────────────────────────────

func main() {
	// Redirect logs to stderr because stdout is reserved for JSON-RPC messages
	log.SetOutput(os.Stderr)
	log.Printf("🔌 %s v%s starting over stdio...", serverName, serverVersion)
	log.Printf("🔗 Backend API: %s", apiURL)

	// Health check on startup (non-blocking)
	go checkHealth()

	// Create scanner with large buffer for handling big payloads
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, maxScannerBuffer), maxScannerBuffer)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // skip empty lines
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("⚠️ Failed to parse JSON-RPC request: %v", err)
			sendError(nil, -32700, "Parse error: invalid JSON", nil)
			continue
		}

		handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Scanner error: %v", err)
	}
}

// ─── Health Check ─────────────────────────────────────────────────────────────

func checkHealth() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL + "/")
	if err != nil {
		log.Printf("⚠️ Backend API health check failed: %v", err)
		log.Printf("💡 Make sure the free-gemini-api Docker container is running on %s", apiURL)
		return
	}
	defer resp.Body.Close()
	log.Printf("✅ Backend API is reachable (status: %d)", resp.StatusCode)
}

// ─── Request Router ───────────────────────────────────────────────────────────

func handleRequest(req JSONRPCRequest) {
	switch req.Method {

	// ── MCP Lifecycle ──
	case "initialize":
		result := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    serverName,
				"version": serverVersion,
			},
		}
		sendResult(req.ID, result)

	// MCP clients send this after handshake — it's a notification, no response needed
	case "notifications/initialized":
		log.Println("✅ Client initialized successfully")
		// JSON-RPC notifications (no ID) must NOT receive a response

	// Keepalive ping from MCP clients
	case "ping":
		sendResult(req.ID, map[string]interface{}{})

	// ── Tool Discovery ──
	case "tools/list":
		sendResult(req.ID, getToolDefinitions())

	// ── Tool Execution ──
	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params: could not parse tool call parameters", nil)
			return
		}

		var args PromptArg
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			sendError(req.ID, -32602, "Invalid arguments: could not parse tool arguments", nil)
			return
		}

		if args.Prompt == "" && params.Name != "reset_session" && params.Name != "check_health" {
			sendResult(req.ID, &CallToolResult{
				Content: []TextContent{{Type: "text", Text: "❌ Error: 'prompt' is required but was empty."}},
				IsError: true,
			})
			return
		}

		result, err := callTool(params.Name, args)
		if err != nil {
			// Return as tool error (isError=true) instead of JSON-RPC error
			// This lets the MCP client display it gracefully
			sendResult(req.ID, &CallToolResult{
				Content: []TextContent{{Type: "text", Text: formatToolError(err)}},
				IsError: true,
			})
			return
		}

		sendResult(req.ID, result)

	default:
		// Unknown methods that look like notifications (no ID) should be silently ignored
		if req.ID == nil {
			log.Printf("📩 Ignoring unknown notification: %s", req.Method)
			return
		}
		sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

// ─── Tool Definitions ─────────────────────────────────────────────────────────

func getToolDefinitions() map[string]interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "generate_image",
				"description": "Generate a high-quality image using Google Gemini's Imagen 3 (Nano Banana 2) from a detailed text prompt. The generated image is automatically cleaned of watermarks. Supports optional Image-to-Image (I2I) by providing a local reference image path.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type":        "string",
							"description": "Detailed text description of the image to generate. Be specific about style, colors, composition, and subject matter for best results.",
						},
						"ref_image_path": map[string]interface{}{
							"type":        "string",
							"description": "Optional: Absolute path to a local reference image for Image-to-Image generation (e.g., /Users/akash/photo.jpg)",
						},
					},
					"required": []string{"prompt"},
				},
			},
			{
				"name":        "generate_video",
				"description": "Generate a short cinematic video (2-10 seconds) using Google Gemini Video from a text prompt. Supports optional Image-to-Video (I2V) by providing a starting image. Video generation may take 1-3 minutes.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type":        "string",
							"description": "Detailed text description of the video to generate. Describe the scene, camera movement, lighting, and subject action.",
						},
						"start_image_path": map[string]interface{}{
							"type":        "string",
							"description": "Optional: Absolute path to a local image to use as the first frame of the video (Image-to-Video)",
						},
					},
					"required": []string{"prompt"},
				},
			},
			{
				"name":        "generate_music",
				"description": "Generate a short music beat or song using YouTube Lyria from a text prompt. Produces high-fidelity audio output.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type":        "string",
							"description": "Detailed description of the music style, mood, instruments, and tempo to generate (e.g., 'lofi piano beat for coding with soft rain ambiance')",
						},
					},
					"required": []string{"prompt"},
				},
			},
			{
				"name":        "chat",
				"description": "Chat with Google Gemini using a text prompt. Supports optional multimodal analysis by attaching a local image or video file for Gemini to analyze, describe, or use as context.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type":        "string",
							"description": "The message or question to send to Gemini",
						},
						"ref_image_path": map[string]interface{}{
							"type":        "string",
							"description": "Optional: Absolute path to a local image to analyze or use as context",
						},
						"ref_video_path": map[string]interface{}{
							"type":        "string",
							"description": "Optional: Absolute path to a local video (.mp4) to analyze or use as context",
						},
					},
					"required": []string{"prompt"},
				},
			},
			{
				"name":        "reset_session",
				"description": "Reset the active chat and generation session to start a fresh conversation.",
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
			{
				"name":        "check_health",
				"description": "Check if the Gemini API backend is healthy and responsive. Returns the connection status.",
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
	}
}

// ─── Tool Execution ───────────────────────────────────────────────────────────

func callTool(toolName string, args PromptArg) (*CallToolResult, error) {
	log.Printf("🔧 Executing tool: %s | prompt: %.80s...", toolName, args.Prompt)

	switch toolName {

	case "generate_image", "generate_video":
		return handleMediaGeneration(toolName, args)

	case "generate_music":
		return handleMusicGeneration(args)

	case "chat":
		return handleChat(args)

	case "reset_session":
		return handleResetSession()

	case "check_health":
		return handleHealthCheck()

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ─── Media Generation (Image & Video) ─────────────────────────────────────────

func handleMediaGeneration(toolName string, args PromptArg) (*CallToolResult, error) {
	var imagePath string
	if toolName == "generate_image" {
		imagePath = args.RefImagePath
	} else {
		imagePath = args.StartImagePath
	}

	// Validate file path if provided
	if imagePath != "" {
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", imagePath)
		}
	}

	var doRequest func() (*http.Response, error)

	if imagePath != "" {
		log.Printf("📎 Using reference file: %s", imagePath)
		doRequest = func() (*http.Response, error) {
			return postMultipart(apiURL+"/chat", args.Prompt, "mcp_client_media", true, imagePath)
		}
	} else {
		doRequest = func() (*http.Response, error) {
			payload := map[string]interface{}{
				"prompt":   args.Prompt,
				"user_id":  "mcp_client_media",
				"new_chat": true, // Always fresh context for media generation
			}
			data, _ := json.Marshal(payload)
			client := &http.Client{Timeout: httpTimeout}
			return client.Post(apiURL+"/chat", "application/json", bytes.NewBuffer(data))
		}
	}

	// Execute with retry
	body, err := executeWithRetry(doRequest)
	if err != nil {
		return nil, err
	}

	var mresp struct {
		Text   string   `json:"text"`
		Images []string `json:"images"`
		Videos []string `json:"videos"`
	}
	if err := json.Unmarshal(body, &mresp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	var msg string
	if toolName == "generate_image" && len(mresp.Images) > 0 {
		msg = fmt.Sprintf("🎨 Image generated successfully!\nURL: %s", mresp.Images[0])
		if len(mresp.Images) > 1 {
			msg += "\n\nAdditional images:"
			for i, img := range mresp.Images[1:] {
				msg += fmt.Sprintf("\n  %d. %s", i+2, img)
			}
		}
	} else if toolName == "generate_video" && len(mresp.Videos) > 0 {
		msg = fmt.Sprintf("🎬 Video generated successfully!\nURL: %s", mresp.Videos[0])
	} else {
		msg = fmt.Sprintf("⚠️ Request processed but no media URL was returned.\nGemini replied: %s", mresp.Text)
	}

	return &CallToolResult{
		Content: []TextContent{{Type: "text", Text: msg}},
	}, nil
}

// ─── Music Generation ─────────────────────────────────────────────────────────

func handleMusicGeneration(args PromptArg) (*CallToolResult, error) {
	doRequest := func() (*http.Response, error) {
		payload := map[string]interface{}{
			"prompt":   args.Prompt,
			"user_id":  "mcp_client_music",
			"new_chat": true,
		}
		data, _ := json.Marshal(payload)
		client := &http.Client{Timeout: httpTimeout}
		return client.Post(apiURL+"/music", "application/json", bytes.NewBuffer(data))
	}

	body, err := executeWithRetry(doRequest)
	if err != nil {
		return nil, err
	}

	var mresp struct {
		Text  string `json:"text"`
		Music []struct {
			LocalPath   string `json:"local_path"`
			DownloadURL string `json:"download_url"`
			Title       string `json:"title"`
		} `json:"music"`
	}
	if err := json.Unmarshal(body, &mresp); err != nil {
		return nil, fmt.Errorf("failed to parse music response: %w", err)
	}

	var msg string
	if len(mresp.Music) > 0 {
		track := mresp.Music[0]
		msg = fmt.Sprintf("🎵 Music generated successfully!\nTitle: %s\nPath: %s", track.Title, track.LocalPath)
		if track.DownloadURL != "" {
			msg += fmt.Sprintf("\nURL: %s", track.DownloadURL)
		}
	} else {
		msg = fmt.Sprintf("⚠️ Music request processed but no track was returned.\nGemini replied: %s", mresp.Text)
	}

	return &CallToolResult{
		Content: []TextContent{{Type: "text", Text: msg}},
	}, nil
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

func handleChat(args PromptArg) (*CallToolResult, error) {
	mediaPath := ""
	if args.RefImagePath != "" {
		mediaPath = args.RefImagePath
	} else if args.RefVideoPath != "" {
		mediaPath = args.RefVideoPath
	}

	// Validate file path if provided
	if mediaPath != "" {
		if _, err := os.Stat(mediaPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", mediaPath)
		}
	}

	var doRequest func() (*http.Response, error)

	if mediaPath != "" {
		log.Printf("📎 Uploading media for chat: %s", mediaPath)
		doRequest = func() (*http.Response, error) {
			return postMultipart(apiURL+"/chat", args.Prompt, "mcp_client_chat", false, mediaPath)
		}
	} else {
		doRequest = func() (*http.Response, error) {
			payload := map[string]interface{}{
				"prompt":   args.Prompt,
				"user_id":  "mcp_client_chat",
				"new_chat": false,
			}
			data, _ := json.Marshal(payload)
			client := &http.Client{Timeout: httpTimeout}
			return client.Post(apiURL+"/chat", "application/json", bytes.NewBuffer(data))
		}
	}

	body, err := executeWithRetry(doRequest)
	if err != nil {
		return nil, err
	}

	var mresp struct {
		Text   string   `json:"text"`
		Images []string `json:"images"`
		Videos []string `json:"videos"`
	}
	if err := json.Unmarshal(body, &mresp); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}

	msg := mresp.Text

	// Append any media URLs that came with the chat response
	if len(mresp.Images) > 0 {
		msg += "\n\n📸 Generated Images:"
		for _, img := range mresp.Images {
			msg += fmt.Sprintf("\n  %s", img)
		}
	}
	if len(mresp.Videos) > 0 {
		msg += "\n\n🎬 Generated Videos:"
		for _, vid := range mresp.Videos {
			msg += fmt.Sprintf("\n  %s", vid)
		}
	}

	return &CallToolResult{
		Content: []TextContent{{Type: "text", Text: msg}},
	}, nil
}

// ─── Reset Session ────────────────────────────────────────────────────────────

func handleResetSession() (*CallToolResult, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	payload := map[string]interface{}{
		"user_id": "mcp_client_chat",
	}
	data, _ := json.Marshal(payload)

	resp, err := client.Post(apiURL+"/reset", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return &CallToolResult{
		Content: []TextContent{
			{Type: "text", Text: "🔄 Session successfully reset! The next messages will start a fresh conversation."},
		},
	}, nil
}

// ─── Health Check Tool ────────────────────────────────────────────────────────

func handleHealthCheck() (*CallToolResult, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(apiURL + "/")
	elapsed := time.Since(start)

	if err != nil {
		return &CallToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("❌ Backend API unreachable: %v\n\n💡 Make sure Docker container 'free-gemini-api' is running:\n   docker compose up -d", err)}},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	msg := fmt.Sprintf("✅ Backend API is healthy\n"+
		"   URL: %s\n"+
		"   Status: %d\n"+
		"   Latency: %dms",
		apiURL, resp.StatusCode, elapsed.Milliseconds())

	return &CallToolResult{
		Content: []TextContent{{Type: "text", Text: msg}},
	}, nil
}

// ─── Retry Logic ──────────────────────────────────────────────────────────────

func executeWithRetry(doRequest func() (*http.Response, error)) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(math.Pow(2, float64(attempt))) * baseRetryWait
			log.Printf("🔄 Retry %d/%d after %v...", attempt+1, maxRetries, wait)
			time.Sleep(wait)
		}

		resp, err := doRequest()
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			// Connection errors are retryable
			if isRetryableConnectionError(err) {
				continue
			}
			return nil, lastErr
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return body, nil
		}

		// Parse error for better messages
		lastErr = parseAPIError(resp.StatusCode, body)

		// Only retry on transient server errors
		if isRetryableStatus(resp.StatusCode) {
			continue
		}

		// Non-retryable errors: return immediately
		return nil, lastErr
	}

	return nil, fmt.Errorf("all %d attempts failed. Last error: %v", maxRetries, lastErr)
}

func isRetryableStatus(status int) bool {
	return status == 500 || status == 502 || status == 503 || status == 504 || status == 429
}

func isRetryableConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "timeout")
}

// ─── Error Formatting ─────────────────────────────────────────────────────────

func parseAPIError(statusCode int, body []byte) error {
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("[%d] %s", statusCode, humanizeError(errResp.Error))
	}
	return fmt.Errorf("[%d] %s", statusCode, string(body))
}

func humanizeError(raw string) string {
	lower := strings.ToLower(raw)

	switch {
	case strings.Contains(lower, "session expired") || strings.Contains(lower, "redirected to login"):
		return "🔑 Session expired! Please refresh cookies:\n" +
			"  1. Open Chrome → gemini.google.com\n" +
			"  2. Login to your Google account\n" +
			"  3. The browser extension will auto-sync cookies"

	case strings.Contains(lower, "cookies.json not found"):
		return "🍪 No session cookies found. Please install & connect the Chrome extension:\n" +
			"  1. Load the 'gemini-extension' folder in chrome://extensions/\n" +
			"  2. Open gemini.google.com and login\n" +
			"  3. The extension will sync cookies automatically"

	case strings.Contains(lower, "rate limit") || strings.Contains(lower, "quota"):
		return "⏳ Rate limit reached. Please wait a few minutes and try again."

	case strings.Contains(lower, "content policy") || strings.Contains(lower, "safety"):
		return "🚫 Content blocked by Gemini safety filters. Try rephrasing your prompt."

	default:
		return raw
	}
}

func formatToolError(err error) string {
	return fmt.Sprintf("❌ Tool execution failed: %v", err)
}

// ─── Multipart Upload ─────────────────────────────────────────────────────────

func postMultipart(targetURL, prompt, userID string, newChat bool, filePath string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add form fields
	if err := w.WriteField("prompt", prompt); err != nil {
		return nil, fmt.Errorf("failed to write prompt field: %w", err)
	}
	if err := w.WriteField("user_id", userID); err != nil {
		return nil, fmt.Errorf("failed to write user_id field: %w", err)
	}
	if newChat {
		if err := w.WriteField("new_chat", "true"); err != nil {
			return nil, fmt.Errorf("failed to write new_chat field: %w", err)
		}
	}

	// Add file if specified
	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file '%s': %w", filePath, err)
		}
		defer file.Close()

		part, err := w.CreateFormFile("image", filepath.Base(filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file part: %w", err)
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, fmt.Errorf("failed to copy file data: %w", err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", targetURL, &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: httpTimeout}
	return client.Do(req)
}

// ─── JSON-RPC Helpers ─────────────────────────────────────────────────────────

func sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	send(resp)
}

func sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}
	send(resp)
}

func send(val interface{}) {
	data, err := json.Marshal(val)
	if err != nil {
		log.Printf("❌ Failed to marshal JSON-RPC message: %v", err)
		return
	}
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
}
