---
name: gemini-music
description: >
  Generate music, songs, and instrumental beats using Google Lyria 3 via
  Gemini. Use when the user wants to create background music, vocal songs
  (Hindi/English/8 languages), lofi beats, cinematic scores, or musical
  transition sweeps. Supports custom lyrics, BPM, key, mood, and genre.
  IMPORTANT: Never use terms like "sound effect" or "SFX" in prompts —
  describe them as musical elements instead to avoid text-only responses.
---

# 🎵 Gemini Music Generation Guide (YouTube Lyria 3)

This guide documents the capabilities, prompting strategies, and integration options for Google Gemini's advanced music generation engine powered by **YouTube Lyria (Lyria 3)**.

---

## 🚀 Overview

Gemini uses the **Lyria 3** family of generative audio models to transform text prompts and reference inputs into high-fidelity, 44.1 kHz stereo audio. Unlike basic MIDI generators, Lyria synthesizes complex arrangements, natural vocal performances, and high-fidelity instruments in a single pass.

---

## 🎛️ Key Capabilities

### 1. Vocals & Lyrics
* **Hindi Support:** Supports vocal generation in 8 languages: **Hindi, English, Spanish, French, German, Japanese, Korean, and Portuguese**.
* **Custom Lyrics:** You can supply your own lyrics (in any supported language), or let the model compose lyrics based on a theme.
* **Singer Demographics:** Specify the vocalist's characteristics (e.g., `"female breathy soprano"`, `"raspy male baritone"`, `"choral backing vocals"`).

### 2. Composition Structure
Lyria can arrange full-length compositions up to ~3 minutes with structural sections:
* **Intro:** Instrumental openings or soft vocal intros.
* **Verse / Chorus:** Classic alternating song structures.
* **Bridge:** Melodic transitions or instrumental solos.
* **Outro / Fade-out:** Clean musical endings.

### 3. Technical Parameters
* **Tempo (BPM):** Specify the exact speed (e.g., `"75 BPM lofi beat"`, `"140 BPM uptempo techno"`).
* **Musical Key:** Choose the scale (e.g., `"C-major"`, `"A-minor"`).
* **Mood:** Guide the emotional tone (e.g., `"nostalgic"`, `"uplifting"`, `"dark cinematic"`, `"melancholic"`).

### 4. Instruments & Genres
* **Instruments:** Acoustic guitar, Grand piano, Synthesizer, Violin, Saxophone, Electric bass, Hip-hop drum kits, etc.
* **Genres:** Lofi hip-hop, Synthwave, Classical cinematic, Acoustic pop, Indie rock, Techno, and Traditional Indian fusion.

---

## ⚠️ Crucial Rule: Sound Effects (SFX) vs. Musical Transitions

The Gemini music generator (YouTube Lyria 3) is strictly designed for **music, melodies, and beats**, not for mechanical Sound Effects (SFX).

*   **Avoid SFX terms**: Do not use words like `"sound effect"`, `"SFX"`, `"mouse click"`, `"foley"`, or `"impact"`. Using these words will trigger Gemini's text safety/assistant filters, causing it to return a text response with instructions/recipes on how to create the sound instead of generating an audio file.
*   **Use musical terms**: To generate transition sweeps, swells, or hits, describe them as **musical elements**:
    *   ❌ *Instead of*: `"Create a whoosh transition sound effect"`
    *   ✅ *Use*: `"A short 2-second ambient synthesizer chord transition with a gradual volume swell, key of C-major"`
    *   ❌ *Instead of*: `"A UI click sound effect"`
    *   ✅ *Use*: `"A single clean staccato synthesizer bell note, high pitch, minimal reverb"`

---

## ✍️ Prompting Strategies & Templates

To get the best results, use structured prompts containing:
`[Genre/Mood] + [Instruments] + [Tempo/Key] + [Vocal Description] + [Theme/Lyrics]`

### 📋 Example Prompt Templates

#### 1. Hindi Vocal Song
> **Prompt**: "A nostalgic indie-pop Hindi song about rain and train journeys. Features a soft acoustic nylon-string guitar and a melancholic flute. 80 BPM. Soft, breathy female soprano vocals singing emotional Hindi lyrics."

#### 2. Lofi Coding Beat (Instrumental)
> **Prompt**: "Relaxing 70 BPM lofi hip-hop study beat. Jazzy electric piano chords, a warm sub-bass, and a dusty vinyl crackle sound effect. Nostalgic and cozy mood. No vocals."

#### 3. High-Energy Synthwave (Sound Effects)
> **Prompt**: "Energetic 120 BPM synthwave track with retro 80s synthesizers, driving retro drum machines, and futuristic lasers sound effects. High tempo, key of G-minor, uplifting cyberpunk ambiance."

#### 4. Cinematic Background Score
> **Prompt**: "Dark cinematic background score for an adventure movie. Orchestral strings, heavy brass hits, slow taiko drums, and dramatic violin solos. Key of D-minor, epic and tense atmosphere."

---

## 🔌 MCP Integration (`generate_music` Tool)

In your AI workflow, you can trigger this capability using the `gemini` MCP tool:

```json
{
  "name": "generate_music",
  "arguments": {
    "prompt": "Your detailed music generation prompt here"
  }
}
```

### Response Format
The tool returns the title, local path of the generated `.mp3` / `.wav` file, and download URL:
```text
🎵 Music generated successfully!
Title: Rain Journey
Path: /app/data/output/music_xxxx.mp3
URL: http://localhost:8002/output/music_xxxx.mp3
```
