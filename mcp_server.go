package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

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

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type PromptArg struct {
	Prompt string `json:"prompt"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type CallToolResult struct {
	Content []TextContent `json:"content"`
}

const apiURL = "http://localhost:8001"

func main() {
	// Redirect logs to stderr because stdout is reserved for JSON-RPC messages
	log.SetOutput(os.Stderr)
	log.Println("🔌 Starting Gemini MCP Server over stdio...")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(req.ID, -32700, "Parse error", nil)
			continue
		}

		handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

func handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		result := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "free-gemini-mcp",
				"version": "1.0.0",
			},
		}
		sendResult(req.ID, result)

	case "tools/list":
		result := map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "generate_image",
					"description": "Generate an image using Imagen 3 from a text prompt",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "Detailed text description of the image to generate",
							},
						},
						"required": []string{"prompt"},
					},
				},
				{
					"name":        "generate_video",
					"description": "Generate a short video using Gemini Video from a text prompt",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "Detailed text description of the video to generate",
							},
						},
						"required": []string{"prompt"},
					},
				},
				{
					"name":        "generate_music",
					"description": "Generate a short music beat/song from a text prompt",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "Detailed description of the style/lofi beat to generate",
							},
						},
						"required": []string{"prompt"},
					},
				},
				{
					"name":        "chat",
					"description": "Chat with Gemini using a text prompt",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "The message to send to Gemini",
							},
						},
						"required": []string{"prompt"},
					},
				},
			},
		}
		sendResult(req.ID, result)

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params", nil)
			return
		}

		var args PromptArg
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			sendError(req.ID, -32602, "Invalid arguments", nil)
			return
		}

		result, err := callTool(params.Name, args.Prompt)
		if err != nil {
			sendError(req.ID, -32603, fmt.Sprintf("Tool call failed: %v", err), nil)
			return
		}

		sendResult(req.ID, result)

	default:
		sendError(req.ID, -32601, "Method not found", nil)
	}
}

func callTool(toolName, prompt string) (*CallToolResult, error) {
	log.Printf("Executing tool %s with prompt: %s", toolName, prompt)

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	switch toolName {
	case "generate_image", "generate_video":
		// Direct chat generation endpoint (with NewChat=true to isolate request)
		payload := map[string]interface{}{
			"prompt":   prompt,
			"user_id":  "mcp_client_" + toolName,
			"new_chat": true,
		}
		data, _ := json.Marshal(payload)

		resp, err := client.Post(apiURL+"/chat", "application/json", bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API error: %s", string(body))
		}

		var mresp struct {
			Text   string   `json:"text"`
			Images []string `json:"images"`
			Videos []string `json:"videos"`
		}
		if err := json.Unmarshal(body, &mresp); err != nil {
			return nil, err
		}

		var msg string
		if toolName == "generate_image" && len(mresp.Images) > 0 {
			msg = fmt.Sprintf("🎨 Image generated successfully!\nURL: %s", mresp.Images[0])
		} else if toolName == "generate_video" && len(mresp.Videos) > 0 {
			msg = fmt.Sprintf("🎬 Video generated successfully!\nURL: %s", mresp.Videos[0])
		} else {
			msg = fmt.Sprintf("⚠️ Request succeeded but no media URL was returned. Reply: %s", mresp.Text)
		}

		return &CallToolResult{
			Content: []TextContent{
				{Type: "text", Text: msg},
			},
		}, nil

	case "generate_music":
		payload := map[string]interface{}{
			"prompt":   prompt,
			"user_id":  "mcp_client_music",
			"new_chat": true,
		}
		data, _ := json.Marshal(payload)

		resp, err := client.Post(apiURL+"/music", "application/json", bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API error: %s", string(body))
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
			return nil, err
		}

		var msg string
		if len(mresp.Music) > 0 {
			msg = fmt.Sprintf("🎵 Music generated successfully!\nTitle: %s\nURL: %s", mresp.Music[0].Title, mresp.Music[0].LocalPath)
		} else {
			msg = fmt.Sprintf("⚠️ Music request succeeded but no track URL was returned. Reply: %s", mresp.Text)
		}

		return &CallToolResult{
			Content: []TextContent{
				{Type: "text", Text: msg},
			},
		}, nil

	case "chat":
		payload := map[string]interface{}{
			"prompt":   prompt,
			"user_id":  "mcp_client_chat",
			"new_chat": false,
		}
		data, _ := json.Marshal(payload)

		resp, err := client.Post(apiURL+"/chat", "application/json", bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API error: %s", string(body))
		}

		var mresp struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(body, &mresp); err != nil {
			return nil, err
		}

		return &CallToolResult{
			Content: []TextContent{
				{Type: "text", Text: mresp.Text},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

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
		log.Printf("Failed to marshal JSON-RPC message: %v", err)
		return
	}
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
}
