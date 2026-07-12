---
name: gemini-visual-analysis
description: >
  Analyze screenshots, UI mockups, and screen recordings using Gemini AI.
  Use when the user wants to debug UI bugs, verify layout alignment, trace
  app flows from a screen recording, or identify visual issues in their
  interface. Also use when the user needs to compress large images (>20MB)
  or heavy videos before uploading for analysis.
---

# 👁️ Gemini Visual Analysis Guide (Images & Videos)

This guide documents the capabilities, prompting strategies, and integration options for using Gemini to perform **visual analysis** on screenshots, mockups, and screen recordings.

---

## 🔌 MCP Integration (`chat` Tool)

To perform visual or video analysis, attach local file paths to the `chat` tool:
*   **For Screenshots/Images:** Use `ref_image_path`
*   **For Screen Recordings/Videos:** Use `ref_video_path`

> [!IMPORTANT]
> **Performance Optimization:** Large high-resolution videos (like 4K screen recordings) and heavy images (> 20MB) are slow to upload and can cause timeouts.
> *   **Videos:** Always compress them to a lower resolution (e.g., 480p) before uploading.
> *   **Images (> 20MB):** Always resize or compress them to under 5MB before uploading. The AI is highly capable of understanding optimized visuals.

### 🎥 Step 1: Compress Video (Local)
Run this command on the host terminal to compress the video to 480p resolution and a lower bitrate:
```bash
ffmpeg -i "/Users/akash/My-work/outputs/original_video.mp4" -vf "scale=-2:480" -vcodec libx264 -crf 28 -acodec aac -ab 128k "/Users/akash/My-work/outputs/compressed_video.mp4"
```

### 🖼️ Step 2: Compress Heavy Images (Local)
Run this built-in macOS command on the host terminal to resize and compress large images to a maximum width of 1920px (reduces size under 5MB):
```bash
sips --resampleWidth 1920 "/Users/akash/My-work/outputs/original_large_image.png" --out "/Users/akash/My-work/outputs/compressed_image.png"
```

### 🔌 Step 3: Send Compressed Media to Gemini MCP
```json
{
  "name": "chat",
  "arguments": {
    "prompt": "Analyze this screen recording showing the login freeze. Tell me what is failing.",
    "ref_video_path": "/Users/akash/My-work/outputs/compressed_video.mp4"
  }
}
```

---

## 🎥 Visual Prompting Templates

### 1. UI Bug & Layout Verification
When a UI element is misaligned, overlapping, or has contrast issues:
> **Prompt**: "Inspect this screenshot of the dashboard UI. Identify any overlapping text, color contrast issues, or misaligned buttons, and write the TailwindCSS/CSS fix to correct it."

### 2. App Flow / Crash Video Analysis
When a feature fails in an interactive flow (e.g. login, payment) and you have a compressed screen recording of the bug:
> **Prompt**: "Analyze this compressed 10-second screen recording showing the login flow failure. Trace the visual transitions step-by-step, explain where the loading spinner freezes, and suggest what frontend/backend synchronization code is failing."
