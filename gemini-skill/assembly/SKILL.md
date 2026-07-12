---
name: gemini-video-assembly
description: >
  Use Gemini AI to analyze media files and generate precise FFmpeg commands
  for intelligent video editing. Use when the user wants to crop/reframe
  videos, duck background music under voiceover, burn subtitles with
  custom styling, or sequence B-roll clips to match narration timing.
  Gemini analyzes the media first, then outputs the exact FFmpeg script.
---

# 🎛️ Gemini-Guided Video Assembly (Intelligent FFmpeg Scripting)

This guide documents how to use Gemini to analyze media files (audio/video) and generate exact, custom **FFmpeg scripts** to perform smart video editing, audio ducking, reframing, and subtitle burning.

---

## 🚀 Overview

While raw command-line tools like `ffmpeg` perform video operations, they lack contextual intelligence (e.g., they don't know where the speaker's face is, or when the voiceover goes silent). 

By sending your media to Gemini for analysis, you can ask the AI to calculate the exact timestamps, crop coordinates, and volume levels, and output the **precise FFmpeg commands** tailored to your files.

---

## 🔌 MCP Integration (`chat` Tool)

Attach your media file (audio or video) to the `chat` tool and ask Gemini to write the custom FFmpeg commands based on its analysis:

```json
{
  "name": "chat",
  "arguments": {
    "prompt": "Listen to this audio track. Tell me the exact timestamps where the speaker is talking, and write the FFmpeg script to duck the background music by 15dB during those parts.",
    "ref_video_path": "/Users/akash/My-work/outputs/voiceover.mp3"
  }
}
```

---

## 📋 Editing & Assembly Prompt Templates

### 1. Smart Cropping & Centering (Landscape 16:9 to Portrait 9:16)
To crop a landscape video to portrait while keeping the main subject in the center of the frame:
> **Prompt**: "Inspect this video. Identify the horizontal coordinate (X-axis) where the main subject is located. Then, write the exact FFmpeg crop command to convert this 16:9 video to a 9:16 portrait video while keeping the subject centered."

### 2. Intelligent Audio Ducking (Music + Voiceover)
To automatically lower background music volume when the narrator speaks:
> **Prompt**: "Analyze the dialogue timestamps in this audio track. Write a custom FFmpeg command using the `sidechain` or `amix` filter to blend my background music file (`bg_music.mp3`) with this voiceover, ducking the music volume down to 15% only when the voiceover is active."

### 3. Smart Subtitle Styling & Burning
To burn subtitles (.srt) onto the video with custom styling matching the video's aesthetic:
> **Prompt**: "Look at the visual style of this video (colors, theme). Choose a matching font, size, and color for subtitles, and write the FFmpeg command using the `subtitles` filter to hardcode my subtitle file (`subtitles.srt`) onto the video with that specific styling."

### 4. B-Roll Timeline Sequencing (Concat & Merge)
To cut and join multiple B-roll clips based on narration pacing:
> **Prompt**: "Listen to this voiceover track. Break down the timing of different sections, and write the FFmpeg concat/merge command to sequence `clip1.mp4`, `clip2.mp4`, and `clip3.mp4` to match the transition points of the audio."
