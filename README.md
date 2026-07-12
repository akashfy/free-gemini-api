![Free Gemini API Banner](assets/banner.png?v=3)

# 🔓 Free Gemini API

High-performance OpenAI-compatible Go server that bridges Google Gemini's web client. Supports text chat, image generation (Imagen 3), video generation, and music synthesis — all with automatic watermark removal.

## ✨ Features

- 💬 **Chat** — `/v1/chat/completions` with SSE streaming
- 🎨 **Image Generation** — `/v1/images/generations` (Imagen 3, auto watermark cleaned)
- 🎬 **Video Generation** — Cinematic 2-10s video outputs
- 🎵 **Music Synthesis** — YouTube Lyria-powered audio beats
- 🔌 **Chrome Cookie-Sync** — Auto session capture via browser extension
- 🤖 **MCP Server** — Built-in stdio MCP with retry logic, health checks & robust error handling

## 🚀 Quick Start

### 1. Run via Docker
```bash
docker compose up -d
```
API runs on `http://localhost:8002`, WebSocket cookie bridge on `9226`.

### 2. Install Chrome Extension
1. Go to `chrome://extensions/` → Enable **Developer mode**
2. **Load unpacked** → select `gemini-extension/` folder
3. Open [gemini.google.com](https://gemini.google.com) and login — cookies sync automatically

### 3. MCP Server (for Cursor/Claude/Antigravity IDE)
```bash
go build -o gemini-mcp mcp_server.go
```
Add to your MCP config:
```json
{
  "mcpServers": {
    "flow-agent": {
      "command": "/path/to/gemini-mcp",
      "args": []
    }
  }
}
```

## 📋 Requirements
- **Docker** (for API server)
- **Chrome** + Extension (for cookie sync)
- **Go 1.21+** (for MCP server compilation)
