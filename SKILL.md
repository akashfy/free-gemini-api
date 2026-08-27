---
name: gemini
description: >-
  Comprehensive unified superpower skill and automation guide for Google Gemini AI.
  Covers conversational chat, coding & wireframe blueprints, Gemini Image generation,
  Gemini Video B-roll, Gemini Music synthesis, audio transcription & SRT subtitles,
  visual & UI bug analysis, smart FFmpeg video assembly & audio ducking, automated Vox-style
  documentary video creation with Remotion, and ultra-fast Needle 2 native tool calling.
---

# ⚡ Google Gemini Unified Superpower Skill Guide

This skill is the **all-in-one superpower toolkit** for Google Gemini, providing complete instructions, architecture workflows, prompt engineering templates, and automated production pipelines across text, code, media generation, Remotion video editing, and native Needle 2 tool calling.

---

## 1. 🏗️ Architecture & Server Overview

- **Go Backend Server:** Running on `http://127.0.0.1:8001`
- **Native Tool Engine:** Needle 2 C++ engine embedded directly via CGO (`~/.cache/cactus-needle/2.0.3/libneedle.dylib`)
- **WebSocket Bridge:** Port `9226` (Automatic cookie capture via Chrome Extension)
- **Model Engine:** `gemini-3.7-flash` (Primary text & vision), `Gemini Image` (Images), `Gemini Video` (Videos), `Gemini Music` (Music)
- **API Endpoints:**
  - `POST /v1/chat/completions` — Drop-in OpenAI-compatible chat & native function/tool calling (SSE streaming supported)
  - `GET /v1/models` — Active model list (`gemini-3.7-flash`)
  - `GET /health` — Real-time server & engine status check
  - `POST /chat` — Unified multipart chat (text, images, screen recordings, ref videos)
  - `POST /images/generations` — Gemini Image generation
  - `POST /music` — Gemini Music synthesis

---

## 2. 🎛️ Capabilities Matrix

| Superpower | Underlying Engine | Key Strengths | Primary Use Case |
|---|---|---|---|
| **Chat & Reasoning** | Gemini 3.7 Flash | 1M+ context window, deep reasoning, structured planning | Brainstorming, logic design, summaries |
| **Code & Blueprints** | Gemini 3.7 Flash | Refactoring, stack trace debugging, SQL schemas | Wireframe-to-HTML/CSS, concurrent Go/JS |
| **Image Generation** | Gemini Image | Photorealistic, 3D isometric, 16:9 / 9:16 / 1:1, I2I | Thumbnails, banners, flat lays, textures |
| **Video B-Roll** | Gemini Video | 8s HD 24fps cinematic footage, joint audio sync | B-roll for coding, nature, cyberpunk scenes |
| **Music Synthesis** | Gemini Music | 44.1kHz stereo audio, 8 languages vocals, custom BPM | Lofi beats, vocal songs, cinematic scores |
| **Audio Transcription** | Gemini Multimodal | Formatted `.srt` / `.vtt`, multi-speaker timestamps | Subtitle generation, podcast transcripts |
| **Visual UI Analysis** | Gemini Vision | Screenshot inspection, layout alignment verification | Debugging UI bugs, screen recording traces |
| **Smart Video Assembly** | Gemini + FFmpeg | Auto-reframing 16:9 -> 9:16, sidechain audio ducking | FFmpeg script synthesis, subtitle burning |
| **Vox Documentary Pipeline** | Remotion + FFmpeg | Automated scene planning, kinetic typography, springs | Complete YouTube documentary video generation |
| **Native Tool Calling** | Needle 2 (45M SLM) | 100% JSON grammar guarantee, 20ms latency, 29MB RAM | OpenAI `tools` dispatch, IoT / Edge routing |

---

## 3. 🎵 Music & Beat Synthesis (Gemini Music)

### Core Rules
- **Vocal Languages:** Supports **Hindi, English, Spanish, French, German, Japanese, Korean, Portuguese**.
- **Song Structure:** Specify sections: `[Intro]`, `[Verse]`, `[Chorus]`, `[Bridge]`, `[Outro]`.
- **BPM & Key:** Always specify exact tempo (e.g. `80 BPM`, `128 BPM`) and key (`C-Major`, `A-Minor`).
- ⚠️ **SFX Avoidance Rule:** Never use words like `"SFX"`, `"sound effect"`, `"mouse click"`, or `"foley"`. Describe them as musical instruments instead (e.g., *"A single high-pitch staccato synth bell note"*).

### Prompt Templates
- **Hindi Vocal Song:** `"A nostalgic indie-pop Hindi song about rain and train journeys. Features soft acoustic nylon guitar and a melancholic flute. 80 BPM. Soft, breathy female soprano vocals singing emotional Hindi lyrics."`
- **Lofi Study Beat:** `"Relaxing 75 BPM lofi hip-hop study beat. Warm Fender Rhodes chords, dusty vinyl texture, smooth sub-bass. Instrumental, no vocals, cozy atmosphere."`
- **Cinematic Orchestral Score:** `"Epic dark cinematic soundtrack for a sci-fi thriller. Low cello drones, heavy brass hits, slow taiko drums, tension-building violin ostinato. Key of D-minor, 100 BPM."`

---

## 4. 🖼️ Image B-Roll Generation (Gemini Image)

### Core Rules
- **Aspect Ratios:** Specify `16:9` (desktop/banners), `9:16` (Shorts/Reels), or `1:1` (square).
- **Structure:** `[Subject/Focus] + [Composition/Framing] + [Lighting & Atmosphere] + [Style/Aesthetics]`
- **Image-to-Image (I2I):** Pass a reference image path to guide lighting and style consistency.

### Prompt Templates
- **Developer Flat Lay:** `"A flat lay top-down view of a modern software engineer's wooden desk. Includes a sleek laptop, mechanical keyboard, wireless mouse, steaming coffee cup, and a green succulent. Soft diffused overhead natural lighting, organized knolling, photorealistic, 16:9 aspect ratio."`
- **Cyberpunk Workspace:** `"Close-up macro shot of a programmer workspace at night. Curved ultrawide monitor displaying glowing cyan code, reflections on desk, neon purple ambient strip lighting. Cinematic depth of field, photorealistic, 16:9 aspect ratio."`
- **3D Isometric AI Concept:** `"A modern 3D isometric render of an AI neural network core. Floating glassmorphic data cubes, golden glowing circuit pathways, clean dark background. Vibrant lighting, 16:9 aspect ratio."`

---

## 5. 🎬 Cinematic Video B-Roll (Gemini Video)

### Core Rules
- **Duration & Frame Rate:** 8-second native HD footage at 24 FPS with synced ambient sound.
- ⚠️ **Video Keyword Rule:** Always start prompt with `"Generate a video of..."` or `"Create a video clip showing..."` to prevent fallback to static Gemini Image.

### Prompt Templates
- **Coding Macro Video:** `"Generate a video of a macro close-up shot showing developer hands typing rapidly on a mechanical keyboard with RGB backlighting. Volumetric neon blue ambient light, shallow depth of field, slow smooth camera pan, 16:9 aspect ratio."`
- **Aerial Drone Landscape:** `"Generate a video of an epic drone aerial sweeping forward over a dense pine forest in the mountains at sunrise. Volumetric golden morning sunrays cutting through the fog, cinematic 24fps, 16:9 aspect ratio."`
- **Vertical Neon Street:** `"Generate a video of a vertical tracking shot following a person walking through a rain-slicked Tokyo alleyway at night. Glowing neon signs reflecting in puddles, cyberpunk aesthetic, smooth forward gimbal motion, 9:16 aspect ratio."`

---

## 6. 📝 Audio Transcription & SRT Subtitles

### Audio Pre-processing Workflow
1. Extract audio from video to minimize upload size:
   ```bash
   ffmpeg -i "input_video.mp4" -vn -acodec libmp3lame -q:a 2 "extracted_audio.mp3"
   ```
2. Send to Gemini with prompt:
   - `"Listen to this audio track and generate a precise SRT subtitle file with clean timestamps (00:00:00,000 --> 00:00:00,000)."`
   - For translation: `"Listen to this Spanish audio, translate to English, and output as an English SRT subtitle file."`

---

## 7. 👁️ Visual & UI Bug Analysis

### Pre-compression Rules (Host-Side)
- **Videos:** Always compress heavy recordings to 480p before upload:
  ```bash
  ffmpeg -i "raw_recording.mp4" -vf "scale=-2:480" -vcodec libx264 -crf 28 "compressed_video.mp4"
  ```
- **Images (>20MB):** Downsample using macOS `sips`:
  ```bash
  sips --resampleWidth 1920 "raw_screenshot.png" --out "compressed_screenshot.png"
  ```

### Analysis Use Cases
- **UI Bug / Misalignment:** Inspect screenshot for overlapping text, color contrast, and output exact CSS/Tailwind fixes.
- **App Crash / Freeze Flow:** Trace screen recording to pinpoint where loading spinners hang and diagnose state mismatch.
- **Wireframe to Code:** Convert UI wireframe screenshots into responsive HTML5 / Vanilla CSS components.

---

## 8. 🎛️ Smart Video Assembly & FFmpeg Scripting

Ask Gemini to analyze media files first, then output exact calculated FFmpeg commands:
1. **Landscape (16:9) to Portrait (9:16) Smart Centering:**
   - Prompt: *"Analyze this video, locate the speaker's X-axis coordinate, and write the FFmpeg crop filter command to convert to 9:16 with the speaker centered."*
2. **Sidechain Audio Ducking (Music + Voice):**
   - Command pattern generated:
     ```bash
     ffmpeg -i voiceover.mp3 -i bg_music.mp3 -filter_complex "[1:a]volume=0.15[bg];[0:a][bg]amix=inputs=2:duration=first" -c:a aac output.mp3
     ```
3. **Smart Subtitle Burning with Aesthetic Styling:**
   - Prompt: *"Select fonts and colors matching this video style and generate the FFmpeg subtitles filter command."*

---

## 9. 🎥 Vox-Style Documentary Video Production Pipeline

### Pipeline Flow
```
Topic → Research (research.json) → Narration Script (script.json) → Scene Breakdown (scene_plan.json)
      → Asset Generator (AI Images, Maps, Textures) → Voiceover (Edge TTS / ElevenLabs)
      → Background Music (Gemini Music) → Sound Effects → Remotion Composition → FFmpeg Final Polish
```

### Voiceover Generation (Edge TTS - Free Local)
```bash
edge-tts --voice "en-US-GuyNeural" --text "What if everything you knew about the ocean was wrong?" --write-media voice.wav
```
*(Recommended: `en-US-GuyNeural` for documentary, `hi-IN-MadhurNeural` for Hindi)*

### Remotion Motion Graphics Rules
- **Snappy Spring Physics:** Use `spring({ frame, fps, config: { damping: 18, stiffness: 120, mass: 0.8 } })`
- **Zero Static Visuals:** Every layer has subtle scale, parallax pan, opacity, or mask reveal.
- **Color Palette:** Warm Paper Background, Vox Yellow Accent (`#FFCC00`), Dark Gray Muted (`#222222`).
- **Typography:** `IBM Plex Sans`, `Inter`, `Helvetica`.

### FFmpeg Final Master Polish
```bash
ffmpeg -i remotion_render.mp4 -vf "eq=contrast=1.05:saturation=1.1,noise=alls=8:allf=t+u" -c:v libx264 -crf 18 -c:a aac -b:a 192k final_documentary.mp4
```

---

## 10. ⚡ Native Needle 2 Tool Calling (Pure Go)

### OpenAI `/v1/chat/completions` Protocol
When a client sends OpenAI `tools` schema, the Go server:
1. Evaluates tools in **<20ms** using embedded Needle 2 C++ engine (`libneedle.dylib`).
2. Returns standard `tool_calls` if a function call is triggered:
   ```json
   {
     "choices": [{
       "finish_reason": "tool_calls",
       "message": {
         "role": "assistant",
         "content": null,
         "tool_calls": [{
           "id": "call_123",
           "type": "function",
           "function": {
             "name": "get_weather",
             "arguments": "{\"city\": \"Delhi\", \"unit\": \"celsius\"}"
           }
         }]
       }
     }]
   }
   ```
3. When client returns `role: "tool"` output, Gemini 3.7 Flash synthesizes the final natural language answer.
