# 🔓 Free Gemini API (Containerized Edition)

A fully containerized, high-performance, OpenAI-compatible Go API server and MCP server that bridges requests to the Gemini Web UI via WebSocket cookie sync.

This version is optimized to run inside **OrbStack / Docker** with **zero local host code dependencies**.

---

## 🚀 How to Run

Start the container in the background:
```bash
docker compose up -d
```

Once started, the following services will be available:
- **HTTP API**: `http://localhost:8001` (OpenAI-compatible endpoints)
- **WebSocket Cookie Bridge**: `ws://localhost:9226` (For Chrome Extension to sync cookies)

---

## 🤖 MCP Server Configuration (Cursor / Claude Desktop)

To use the built-in MCP server directly inside Cursor or Claude Desktop, configure it to run through the Docker container:

```json
{
  "mcpServers": {
    "free-gemini": {
      "command": "docker",
      "args": ["exec", "-i", "free-gemini-api", "/app/gemini-mcp"]
    }
  }
}
```

---

## 📁 Persistent Volumes

All temporary files, outputs, and active session cookies are stored inside the named Docker volume:
- `gemini-data` (mounted inside the container at `/app/data`)
