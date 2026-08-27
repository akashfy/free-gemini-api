# 🔓 Free Gemini 3.7 API (with Native Needle 2 Tool Calling)

High-performance OpenAI-compatible Go server that bridges Google Gemini 3.7 Flash web client with **Needle 2 Native C++ Tool Calling Engine**. Supports text chat, tool calling, Gemini Image generation, cinematic video generation, and music synthesis.

---

## ✨ Key Features

- 💬 **OpenAI Chat Completions** — `/v1/chat/completions` with SSE streaming
- ⚡ **Native Tool Calling (Needle 2)** — 100% grammar-constrained JSON function calling in <20ms (Runs on 29MB RAM via native C++ engine)
- 🎨 **Image Generation** — `/images/generations` (Gemini Image with auto-watermark removal)
- 🎬 **Video Generation** — `/chat` (Gemini Video cinematic 8-second video outputs)
- 🎵 **Music Synthesis** — `/music` (Gemini Music powered audio tracks)
- 🔌 **Chrome Cookie-Sync** — Auto session capture via browser extension (WebSocket on port `9226`)
- 🤖 **MCP Server** — Built-in stdio MCP server for Cursor, Windsurf, Claude, and Antigravity IDE

---

## 🚀 Quick Start

### 1. Run the Server
```bash
go run main.go
```
* API Server: `http://localhost:8001`
* Cookie WebSocket Bridge: `ws://localhost:9226`

### 2. Install Chrome Extension (One-time Setup)
1. Open Chrome and navigate to `chrome://extensions/`
2. Enable **Developer mode** (top right toggle)
3. Click **Load unpacked** and select the `gemini-extension/` folder inside this repository
4. Open [gemini.google.com](https://gemini.google.com) and log in — cookies sync automatically to `cookies/cookies.json`

### 3. Test OpenAI Endpoint with Tool Calling
```bash
curl -X POST http://localhost:8001/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [{"role": "user", "content": "What is the weather in Delhi right now in celsius?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "parameters": {
          "type": "object",
          "properties": {
            "city": {"type": "string"},
            "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
          },
          "required": ["city"]
        }
      }
    }]
  }'
```

---

## 🔌 MCP Server Setup (for Cursor / Antigravity IDE)

You can run the MCP server directly using the `--mcp` flag:

```bash
go run main.go --mcp
```

Or add directly to your IDE's `mcp_config.json`:
```json
{
  "mcpServers": {
    "gemini": {
      "command": "/absolute/path/to/free-gemini-api/main",
      "args": ["--mcp"]
    }
  }
}
```

---

## 📋 Requirements
- **Go 1.21+**
- **Chrome** + `gemini-extension` (for cookie synchronization)
