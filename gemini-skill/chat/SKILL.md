---
name: gemini-chat
description: >
  Use Gemini for conversational reasoning, brainstorming, planning, and
  general text-based queries. Use when the user wants to outline workflows,
  prepare design proposals, brainstorm product names, summarize text,
  practice interviews, or answer general knowledge questions.
---

# 💬 Gemini Conversational Chat Guide

This guide documents the capabilities, prompting strategies, and integration options for using Gemini for **conversational reasoning, brainstorming, planning, and general text-based queries**.

---

## 🔌 MCP Integration (`chat` Tool)

Invoke the conversational chat tool via the `gemini` MCP server with simple text queries:

```json
{
  "name": "chat",
  "arguments": {
    "prompt": "What are the key advantages of using WebSockets over HTTP long polling?"
  }
}
```

---

## 💡 Chat & Brainstorming Templates

### 1. Planning & Outlining Workflows
To plan a new feature or outline development steps:
> **Prompt**: "I want to build a real-time multiplayer whiteboard app. Outline the technical steps, backend synchronization logic, and key features I need to plan for the initial MVP."

### 2. Design Proposals & Copywriting
To write proposals, emails, or names:
> **Prompt**: "Brainstorm 10 catchy and professional names for a Go-based watermark-removal tool. Explain the reasoning behind each suggestion."

### 3. Summarization & Explanation
To condense long documents or explain concepts simply:
> **Prompt**: "Explain how Docker container networking works and how port mapping connects the host to the container. Explain it simply like I'm five."

### 4. Interview & Skill Testing
To practice interviews or test concepts:
> **Prompt**: "Ask me 5 multiple-choice questions about Go concurrency patterns (goroutines, channels, mutexes) one by one to test my understanding."
