package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"goapi/gemini"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func FormatOpenAIToolsForNeedle(tools []map[string]any) string {
	var needleTools []map[string]any
	for _, t := range tools {
		if t["type"] == "function" && t["function"] != nil {
			if fn, ok := t["function"].(map[string]any); ok {
				needleTools = append(needleTools, map[string]any{
					"name":        fn["name"],
					"description": fn["description"],
					"parameters":  fn["parameters"],
				})
			}
		} else {
			needleTools = append(needleTools, t)
		}
	}
	bytes, _ := json.Marshal(needleTools)
	return string(bytes)
}

// HandleUnifiedChat handles /chat endpoint (text, images, screen recordings, ref videos)
func HandleUnifiedChat(c fiber.Ctx) error {
	var prompt, userID string
	var newChat, stream bool
	var images []gemini.ImageInput

	contentType := string(c.Request().Header.ContentType())

	if strings.Contains(contentType, "multipart/form-data") {
		prompt = c.FormValue("prompt")
		userID = c.FormValue("user_id")
		newChat = c.FormValue("new_chat") == "true"
		stream = c.FormValue("stream") == "true"

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid multipart form"})
		}

		var fileHeaders []*multipart.FileHeader
		if fhs, ok := form.File["images"]; ok {
			fileHeaders = append(fileHeaders, fhs...)
		}
		if fhs, ok := form.File["image"]; ok {
			fileHeaders = append(fileHeaders, fhs...)
		}

		for _, fh := range fileHeaders {
			file, err := fh.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				continue
			}

			mime := fh.Header.Get("Content-Type")
			if mime == "" {
				ext := strings.ToLower(fh.Filename)
				switch {
				case strings.HasSuffix(ext, ".png"):
					mime = "image/png"
				case strings.HasSuffix(ext, ".webp"):
					mime = "image/webp"
				case strings.HasSuffix(ext, ".gif"):
					mime = "image/gif"
				case strings.HasSuffix(ext, ".mp4"):
					mime = "video/mp4"
				default:
					mime = "image/jpeg"
				}
			}
			images = append(images, gemini.ImageInput{Data: data, Filename: fh.Filename, MimeType: mime})
		}
	} else {
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
		UserSessions.Delete(sessionID)
	}

	client, err := GetOrCreateClient(sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var resp *gemini.GeminiResponse

	if stream && len(images) == 0 {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		return c.SendStreamWriter(func(w *bufio.Writer) {
			streamResp, streamErr := client.AskStream(prompt, func(chunk string) {
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

			if len(streamResp.Images) > 0 {
				for i, imgURL := range streamResp.Images {
					filename := fmt.Sprintf("img_%s_%d.png", streamResp.ResponseID, i)
					if streamResp.ResponseID == "" {
						filename = fmt.Sprintf("img_%d_%d.png", time.Now().Unix(), i)
					}
					if err := DownloadAndClean(client, imgURL, filename, "image"); err == nil {
						streamResp.Images[i] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
					}
				}
			}

			if len(streamResp.Videos) > 0 {
				vidURL := streamResp.Videos[0]
				if vidURL != "" && len(vidURL) > 50 {
					filename := fmt.Sprintf("vid_%s_0.mp4", streamResp.ResponseID)
					if streamResp.ResponseID == "" {
						filename = fmt.Sprintf("vid_%d_0.mp4", time.Now().Unix())
					}
					if err := DownloadAndClean(client, vidURL, filename, "video"); err == nil {
						streamResp.Videos[0] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
					}
				}
			}

			final, _ := json.Marshal(streamResp)
			fmt.Fprintf(w, "data: %s\n\n", final)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			w.Flush()
		})
	}

	if len(images) > 0 {
		log.Printf("📸 Image/Media request: %d items, prompt: %s", len(images), prompt)
	}

	resp, err = ExecuteWithFailover(sessionID, func(cl *gemini.GeminiClient) (*gemini.GeminiResponse, error) {
		if len(images) > 0 {
			return cl.AskWithImages(prompt, images)
		}
		return cl.Ask(prompt)
	})

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			UserSessions.Delete(sessionID)
			return c.Status(504).JSON(fiber.Map{"error": "Request timed out. Session reset."})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(resp.Images) > 0 {
		for i, imgURL := range resp.Images {
			filename := fmt.Sprintf("img_%s_%d.png", resp.ResponseID, i)
			if resp.ResponseID == "" {
				filename = fmt.Sprintf("img_%d_%d.png", time.Now().Unix(), i)
			}
			if err := DownloadAndClean(client, imgURL, filename, "image"); err == nil {
				resp.Images[i] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
			}
		}
	}

	if len(resp.Videos) > 0 {
		vidURL := resp.Videos[0]
		if vidURL != "" && len(vidURL) > 50 {
			filename := fmt.Sprintf("vid_%s_0.mp4", resp.ResponseID)
			if resp.ResponseID == "" {
				filename = fmt.Sprintf("vid_%d_0.mp4", time.Now().Unix())
			}
			if err := DownloadAndClean(client, vidURL, filename, "video"); err == nil {
				resp.Videos[0] = fmt.Sprintf("http://%s/output/%s", c.Host(), filename)
			}
		}
	}

	return c.JSON(resp)
}

// HandleOpenAIChatCompletions handles /v1/chat/completions with Needle 2 Native Tool Calling
func HandleOpenAIChatCompletions(c fiber.Ctx) error {
	var req OpenAIChatCompletionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if len(req.Messages) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Messages array is empty"})
	}

	model := req.Model
	if model == "" {
		model = "gemini-3.7-flash"
	}

	var prompt string
	hasToolResults := false
	var fullPromptBuilder strings.Builder

	for _, m := range req.Messages {
		contentStr := ""
		if m.Content != nil {
			contentStr = *m.Content
		}
		if m.Role == "user" {
			prompt = contentStr
			fullPromptBuilder.WriteString("User: " + contentStr + "\n")
		} else if m.Role == "assistant" {
			if len(m.ToolCalls) > 0 {
				tcJSON, _ := json.Marshal(m.ToolCalls)
				fullPromptBuilder.WriteString("Assistant called tools: " + string(tcJSON) + "\n")
			} else if contentStr != "" {
				fullPromptBuilder.WriteString("Assistant: " + contentStr + "\n")
			}
		} else if m.Role == "tool" {
			hasToolResults = true
			fullPromptBuilder.WriteString(fmt.Sprintf("[Tool Result for %s]: %s\n", m.Name, contentStr))
		} else if m.Role == "system" {
			fullPromptBuilder.WriteString("[System]: " + contentStr + "\n")
		}
	}

	if prompt == "" && len(req.Messages) > 0 {
		lastM := req.Messages[len(req.Messages)-1]
		if lastM.Content != nil {
			prompt = *lastM.Content
		}
	}

	// STEP 1: If tools provided and not yet executed, evaluate with Needle 2 C++ engine
	if len(req.Tools) > 0 && !hasToolResults && prompt != "" {
		engine, nErr := gemini.GetNeedleEngine()
		if nErr == nil {
			toolsJSON := FormatOpenAIToolsForNeedle(req.Tools)
			needleResp, cErr := engine.CompleteTools(prompt, toolsJSON)
			if cErr == nil && needleResp != nil && needleResp.Type == "call" && len(needleResp.FunctionCalls) > 0 {
				var toolCalls []OpenAIToolCall
				for _, fc := range needleResp.FunctionCalls {
					argsBytes, _ := json.Marshal(fc.Arguments)
					toolCalls = append(toolCalls, OpenAIToolCall{
						ID:   "call_" + uuid.NewString()[:8],
						Type: "function",
						Function: OpenAIToolCallFunction{
							Name:      fc.Name,
							Arguments: string(argsBytes),
						},
					})
				}

				return c.JSON(OpenAIChatCompletionResponse{
					ID:      "chatcmpl-" + uuid.NewString()[:12],
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []OpenAIChoice{
						{
							Index: 0,
							Message: OpenAIChatMessage{
								Role:      "assistant",
								Content:   nil,
								ToolCalls: toolCalls,
							},
							FinishReason: "tool_calls",
						},
					},
				})
			}
		} else {
			log.Printf("⚠️ Needle engine load notice: %v", nErr)
		}
	}

	sessionID := c.IP()
	client, err := GetOrCreateClient(sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	askPrompt := prompt
	if hasToolResults {
		fullPromptBuilder.WriteString("Assistant: ")
		askPrompt = fullPromptBuilder.String()
	}

	if req.Stream {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		return c.SendStreamWriter(func(w *bufio.Writer) {
			chunkID := "chatcmpl-" + uuid.NewString()[:12]
			createdTime := time.Now().Unix()

			_, streamErr := client.AskStream(askPrompt, func(chunk string) {
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

	resp, err := ExecuteWithFailover(sessionID, func(cl *gemini.GeminiClient) (*gemini.GeminiResponse, error) {
		return cl.Ask(askPrompt)
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			UserSessions.Delete(sessionID)
			return c.Status(504).JSON(fiber.Map{"error": "Request timed out. Session reset."})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	respContent := resp.Text
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
					Content: &respContent,
				},
				FinishReason: "stop",
			},
		},
	}

	return c.JSON(openAIResp)
}
