---
name: gemini-code
description: >
  Use Gemini for code generation, refactoring, debugging stack traces,
  database schema design, and converting wireframe images into frontend
  code. Use when the user wants to write new functions, fix terminal
  errors, design SQL schemas, generate documentation, or convert a
  UI mockup/wireframe into HTML/CSS code.
---

# 💻 Gemini Coding & Development Guide

This guide documents the capabilities, prompting strategies, and integration options for using Gemini to write code, refactor applications, debug stack traces, design databases, and convert visual mockups/wireframes into code.

---

## 🔌 MCP Integration (`chat` Tool)

To perform coding tasks or generate templates, use the `chat` tool from the `gemini` MCP server:

```json
{
  "name": "chat",
  "arguments": {
    "prompt": "Write a robust Go function to download files concurrently using a worker pool."
  }
}
```

---

## 📝 Coding Templates & Use Cases

### 1. Code Generation & Refactoring
To write new functions or refactor legacy code:
> **Prompt**: "Write a clean and robust Javascript function to debounce user input. Include inline documentation and edge-case handling (like immediate execution options)."

### 2. Terminal Errors & Stack Trace Debugging
To resolve compiler errors or runtime panics:
> **Prompt**: "I got this error in my terminal: [Paste terminal error logs/stack trace here]. Explain what caused this issue and provide the exact code modification to fix it."

### 3. Database Schema & Architecture Design
To design robust relational schemas or microservice architectures:
> **Prompt**: "Design a PostgreSQL schema for a subscription and billing system. Include tables for users, plans, subscriptions, and transaction logs with correct relational foreign keys."

### 4. Wireframe-to-Code Blueprinting
To convert a UI design mockup or wireframe image into frontend code:
> **Prompt**: "Analyze this wireframe layout and generate the corresponding responsive HTML5 and Vanilla CSS code to build this interface exactly as pictured."
*(Note: Provide the wireframe image path using `ref_image_path` in the chat tool arguments.)*

### 5. Documentation & Code Explanations
To explain code or auto-generate docstrings:
> **Prompt**: "Analyze this function and explain its logic step-by-step: [Paste code here]. Write a clean JSDoc/Go docstring for it."
