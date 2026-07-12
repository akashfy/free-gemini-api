---
name: gemini-video-broll
description: >
  Generate cinematic B-roll video clips using Google Veo 3.1 via Gemini.
  Use when the user wants to create short video clips for editing, such as
  coding scenes, tech workspace shots, nature aerials, or cyberpunk streets.
  Supports 16:9 and 9:16 aspect ratios with synchronized ambient audio.
  IMPORTANT: Always start prompts with "Generate a video of..." to avoid
  Imagen 3 static image fallback.
---

# 🎬 Gemini B-Roll Video Generation Guide (Google Veo 3.1)

This guide documents the capabilities, prompting strategies, and integration options for generating high-quality B-roll video clips using Google Gemini's advanced video model (**Google Veo 3.1**).

---

## 🚀 Overview & Specifications

Google Veo 3.1 is a state-of-the-art generative video model capable of outputting cinematic, high-definition footage tailored for professional B-roll editing.

*   **Duration:** Native base duration of **8 seconds** (4s and 6s also supported).
*   **Aspect Ratios:** Supports **16:9 (Landscape)** for wide desktop screens and **9:16 (Portrait)** for mobile Reels/Shorts.
*   **Resolution:** Generates clean, crisp footage in HD/1080p (and up to 4K on supported interfaces).
*   **Frame Rate:** Rendered at a cinematic **24 FPS**.
*   **Joint Audio Generation:** Veo 3.1 natively generates **synchronized ambient audio and sound effects** directly embedded in the video file.

---

## ⚠️ Crucial Rule: Avoid Static Image Fallback

Because Gemini supports both image generation (Imagen 3) and video generation (Veo 3.1) in its model ecosystem, **you must explicitly instruct the model to generate a video**.

*   **Explicit Action Verbs:** Always start your prompt with phrases like `"Generate a video of..."`, `"Create a short video clip showing..."`, or `"A video showing..."`.
*   **Avoid purely static descriptions:** If a prompt only describes visual details (e.g., `"A close-up shot of a keyboard, 16:9 ratio"`) without containing the word `"video"`, Gemini will default to generating a static image instead of a video.

---

## 🎥 Cinematic Prompting Guide

To generate professional B-roll, structure your prompt with these elements:
`[Subject/Action] + [Camera Movement & Framing] + [Lighting & Atmosphere] + [Aspect Ratio / Style]`

### 1. Camera Movements & Framing
*   **Macro / Close-up:** Focuses on tiny details (e.g., keyboard keys, wiring, water droplets).
*   **Drone / Aerial:** High-angle sweeping shots (e.g., city skyline, mountain peaks).
*   **Panning:** Camera rotates horizontally.
*   **Tracking / Dolly:** Camera moves smoothly alongside the subject.
*   **Orbit:** Camera circles 360-degrees around the subject.

### 2. Lighting & Atmosphere
*   **Cinematic Golden Hour:** Warm, soft, low-angle natural sunlight.
*   **Volumetric / Moody:** Rays of light cutting through dust or mist (great for mystery/coding environments).
*   **Neon / Cyberpunk:** Vibrant pink, purple, and blue light sources.
*   **Soft Diffused:** Clean, shadowless lighting (great for product showcases).

---

## 📋 Example B-Roll Prompt Templates

### 1. Developer / Coding B-Roll (Keyboard Close-up)
> **Prompt**: "A macro close-up shot of a developer's hands typing rapidly on a mechanical keyboard with glowing RGB backlighting. Volumetric neon blue and purple ambient lighting in a dark room. Slow panning shot, shallow depth of field, photorealistic, 16:9 aspect ratio."

### 2. Tech Workspace B-Roll (Sweeping Desk Shot)
> **Prompt**: "A smooth dolly tracking shot sweeping across a modern minimalist wooden desk. Shows a laptop displaying glowing lines of green code, a hot steaming cup of coffee next to it, and a tiny green desk plant. Warm golden hour lighting filtering through window blinds, 16:9 aspect ratio."

### 3. Cinematic Nature B-Roll (Drone Landscape)
> **Prompt**: "An epic aerial drone shot panning over a foggy, dense pine forest at sunrise. Volumetric sunrays cutting through the morning mist, lush green and golden tones. 24 FPS, cinematic, high-resolution, 16:9 aspect ratio."

### 4. Vertical Mobile B-Roll (Neon Cyberpunk Street)
> **Prompt**: "A vertical tracking shot of a person walking through a rain-slicked alleyway in Tokyo. Glowing neon signs reflecting in puddles on the ground. Cyberpunk aesthetic, cinematic camera movement, 9:16 aspect ratio."

---

## 🔌 MCP Integration (`generate_video` Tool)

Invoke the video generation tool via the `gemini` MCP server:

```json
{
  "name": "generate_video",
  "arguments": {
    "prompt": "Your detailed cinematic B-roll prompt here"
  }
}
```

### Response Format
The tool returns the title, local container path, and download URL of the generated `.mp4` file:
```text
🎬 Video generated successfully!
Title: Cyberpunk Coding
Path: /app/data/output/video_xxxx.mp4
URL: http://localhost:8002/output/video_xxxx.mp4
```

> [!NOTE]
> Like the music tool, any generated video file can be copied or downloaded from the HTTP URL directly to your workspace outputs folder using a post-generation curl.
