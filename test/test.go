package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "http://localhost:8001"
const userID = "test_go_client"

type ChatRequest struct {
	Prompt  string `json:"prompt"`
	UserID  string `json:"user_id"`
	NewChat bool   `json:"new_chat"`
	Stream  bool   `json:"stream"`
}

type GeminiResponse struct {
	Text           string      `json:"text"`
	ConversationID string      `json:"conversation_id"`
	ResponseID     string      `json:"response_id"`
	ChoiceID       string      `json:"choice_id"`
	Images         []string    `json:"images"`
	Videos         []string    `json:"videos"`
	Music          interface{} `json:"music"`
	Elapsed        float64     `json:"elapsed"`
}

func main() {
	fmt.Println("🚀 Starting API Server Endpoints Verification Tests...")
	fmt.Printf("🔗 Target API Base URL: %s\n\n", baseURL)

	// Test 1: Simple Chat Request
	runTextTest()

	// Test 2: Image Generation (Imagen 3)
	runImageTest()

	// Test 3: Music/Song Generation
	runMusicTest()

	// Test 4: Video Generation (Gemini Video)
	runVideoTest()

	fmt.Println("\n🏁 All tests finished!")
}

func runTextTest() {
	fmt.Println("==================================================")
	fmt.Println("📝 Test 1: Text Chat Generation")
	fmt.Println("==================================================")

	payload := ChatRequest{
		Prompt:  "Who are you? Answer in exactly one short sentence.",
		UserID:  userID,
		NewChat: true,
	}

	resp, err := sendPost("/chat", payload)
	if err != nil {
		fmt.Printf("❌ Text Chat test failed: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Status: SUCCESS\n")
	fmt.Printf("💬 Reply: %s\n", resp.Text)
	fmt.Printf("⏱️  Time Elapsed: %.2f seconds\n\n", resp.Elapsed)
}

func runImageTest() {
	fmt.Println("==================================================")
	fmt.Println("🎨 Test 2: Image Generation (Imagen 3)")
	fmt.Println("==================================================")
	fmt.Println("⏳ Please wait, image generation and watermark cleaning takes a moment...")

	payload := ChatRequest{
		Prompt:  "Generate a beautiful minimalist logo of a glowing purple neon butterfly.",
		UserID:  userID,
		NewChat: true,
	}

	resp, err := sendPost("/chat", payload)
	if err != nil {
		fmt.Printf("❌ Image Gen test failed: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Status: SUCCESS\n")
	if len(resp.Images) > 0 {
		fmt.Printf("📸 Generated Image URL: %s\n", resp.Images[0])
	} else {
		fmt.Println("⚠️  Response text returned but no image URL found in response.")
		fmt.Printf("💬 Response: %s\n", resp.Text)
	}
	fmt.Printf("⏱️  Time Elapsed: %.2f seconds\n\n", resp.Elapsed)
}

func runMusicTest() {
	fmt.Println("==================================================")
	fmt.Println("🎵 Test 3: Music/Song Generation")
	fmt.Println("==================================================")
	fmt.Println("⏳ Generating track...")

	payload := ChatRequest{
		Prompt:  "Generate a short lofi piano beat for coding.",
		UserID:  userID,
		NewChat: true,
	}

	resp, err := sendPost("/music", payload)
	if err != nil {
		fmt.Printf("❌ Music test failed: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Status: SUCCESS\n")
	fmt.Printf("💬 Message: %s\n", resp.Text)
	
	// Print raw music details
	musicData, _ := json.MarshalIndent(resp.Music, "", "  ")
	fmt.Printf("🎶 Music Data: %s\n", string(musicData))
	fmt.Printf("⏱️  Time Elapsed: %.2f seconds\n\n", resp.Elapsed)
}

func runVideoTest() {
	fmt.Println("==================================================")
	fmt.Println("🎬 Test 4: Video Generation (Gemini Video)")
	fmt.Println("==================================================")
	fmt.Println("💡 Note: Video generation consumes significant quota. Running test now...")

	payload := ChatRequest{
		Prompt:  "Generate a 2-second cinematic video of waves gently washing onto a sandy beach.",
		UserID:  userID,
		NewChat: true,
	}

	resp, err := sendPost("/chat", payload)
	if err != nil {
		fmt.Printf("❌ Video Gen test failed: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Status: SUCCESS\n")
	if len(resp.Videos) > 0 {
		fmt.Printf("📹 Generated Video URL: %s\n", resp.Videos[0])
	} else {
		fmt.Println("⚠️  No video URL returned (may have reached account daily limit/quota).")
		fmt.Printf("💬 Response: %s\n", resp.Text)
	}
	fmt.Printf("⏱️  Time Elapsed: %.2f seconds\n\n", resp.Elapsed)
}

func sendPost(endpoint string, payload interface{}) (*GeminiResponse, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 6 * time.Minute, // High timeout for media generation
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result GeminiResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w (body: %s)", err, string(bodyBytes))
	}

	// Clean raw formatting of contribution URL in response if needed
	if len(result.Images) > 0 {
		result.Images[0] = strings.ReplaceAll(result.Images[0], `\`, "")
	}

	return &result, nil
}
