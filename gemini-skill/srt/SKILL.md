---
name: gemini-srt-transcription
description: >
  Transcribe audio files and generate SRT/VTT subtitle files using Gemini.
  Use when the user wants to create subtitles from audio or video, transcribe
  podcasts/meetings/interviews, or translate and subtitle speech into another
  language. Always extract audio first via FFmpeg before sending to Gemini.
---

# 📝 Gemini Audio Transcription & SRT Subtitle Guide

This guide documents the capabilities, prompting strategies, and integration options for using Gemini to **transcribe audio tracks** and **generate subtitle (.srt/.vtt) files** at high speed.

---

## 🚀 Overview

Gemini can ingest audio files (such as `.mp3`, `.wav`, or `.m4a`) or video files with sound to transcribe speech, translate dialogue, and output formatted subtitle files with precise time stamps.

---

## 🔌 MCP Integration (`chat` Tool)

To ensure fast uploads and prevent API processing timeouts, **always provide a direct audio file** (`.mp3`, `.wav`, or `.m4a`) instead of a full video file.

If you have a local video file, you must first extract the audio track using `ffmpeg` before sending it to Gemini.

### 🎥 Step 1: Extract Audio from Video (Local)
Run this command on the host terminal to extract the audio track as an MP3:
```bash
ffmpeg -i "/Users/akash/My-work/outputs/video_file.mp4" -vn -acodec libmp3lame -q:a 2 "/Users/akash/My-work/outputs/extracted_audio.mp3"
```

### 🔌 Step 2: Send Audio to Gemini MCP
Pass the path of the extracted audio file to the `ref_video_path` parameter (which acts as the media input channel):
```json
{
  "name": "chat",
  "arguments": {
    "prompt": "Listen to this audio track and generate a formatted SRT subtitle file with timestamps.",
    "ref_video_path": "/Users/akash/My-work/outputs/extracted_audio.mp3"
  }
}
```

---

## 📋 Transcription & SRT Templates

### 1. Generating SRT Subtitles
To extract speech and format it as subtitles for video players:
> **Prompt**: "Listen to this audio file and generate a precise SRT subtitle file. Make sure the timestamps format is correct (00:00:00,000 --> 00:00:00,000) and the text is clean."

### 2. Audio Transcription (Full Text)
To convert a podcast, meeting, or interview recording into readable text:
> **Prompt**: "Transcribe this audio file completely. Separate the speakers if possible (Speaker A, Speaker B) and clean up any filler words (like 'um', 'uh')."

### 3. Subtitle Translation
To transcribe and translate speech to another language:
> **Prompt**: "Listen to this Spanish audio file, translate it to English, and output the result directly as a formatted English SRT subtitle file."
