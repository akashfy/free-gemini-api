# 🔓 Free Gemini API — Docker Setup

Containerized Go API server running inside **OrbStack / Docker** with zero local dependencies.

---

## 🚀 Run

```bash
docker compose up -d
```

| Service | URL |
|---|---|
| HTTP API | `http://localhost:8002` |
| WebSocket Cookie Bridge | `ws://localhost:9226` |

---

## 🤖 MCP Server Config

**Option 1 — Local binary (Recommended):**
```json
{
  "mcpServers": {
    "flow-agent": {
      "command": "/Users/akash/My-work/free-gemini-api/gemini-mcp",
      "args": []
    }
  }
}
```

**Option 2 — Via Docker exec:**
```json
{
  "mcpServers": {
    "flow-agent": {
      "command": "docker",
      "args": ["exec", "-i", "free-gemini-api", "/app/gemini-mcp"]
    }
  }
}
```

---

## 📁 Volumes

- `gemini-data` → mounted at `/app/data` (cookies, outputs, temp files)
- `/Users` → mounted read-only for local file access (image/video uploads)
