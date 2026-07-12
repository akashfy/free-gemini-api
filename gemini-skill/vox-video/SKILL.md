---
name: gemini-vox-video
description: >
  Generate premium Vox-style documentary videos automatically using a full
  production pipeline. Use when the user wants to create educational or
  documentary videos in the style of Vox, Johnny Harris, MagnatesMedia,
  RealLifeLore, or Wendover Productions. Covers research, scripting, scene
  planning, AI asset generation (images, maps, infographics), voiceover,
  music, SFX, Remotion motion graphics composition, and FFmpeg final polish.
  NOT for simple slideshows — every frame must have motion and cinematic quality.
---

# VOX Documentary Video Generator

Version: 1.0

## Objective

Generate premium Vox-style documentary videos automatically.

The system must create every asset from scratch and assemble everything using
Remotion + FFmpeg.

Final output must look comparable to:

- Vox
- Johnny Harris
- MagnatesMedia
- RealLifeLore
- Wendover Productions

NOT like slideshow videos.

---

# Pipeline

```
Topic
  ↓
Research
  ↓
Script
  ↓
Scene Breakdown
  ↓
Asset Planning
  ↓
Generate Assets
  ↓
Voiceover
  ↓
Music
  ↓
SFX
  ↓
Motion Graphics
  ↓
Remotion Rendering
  ↓
FFmpeg Polish
  ↓
Final Export
```

---

# Step 1: Research

Collect:

- Facts
- Timeline
- Maps
- Statistics
- Locations
- Key Events
- Quotes

Output: `research.json`

---

# Step 2: Script

Generate narration.

### Rules

- Conversational
- Curiosity driven
- No fluff
- Short sentences
- Hook every 5-8 seconds

Output: `script.json`

### Example

**Scene 01**

| Field | Value |
|---|---|
| Narration | What if everything you knew about the ocean was wrong? |
| Duration | 7s |

---

# Step 3: Scene Planning

Every narration block becomes one visual scene.

Generate: `scene_plan.json`

Each scene contains:

- duration
- camera
- transition
- assets
- overlays
- animations

---

# Step 4: Asset Generator

Generate automatically.

### Assets include

- ✓ AI Images
- ✓ Maps
- ✓ Icons
- ✓ Infographics
- ✓ Backgrounds
- ✓ Newspaper
- ✓ Paper textures
- ✓ Arrows
- ✓ Charts
- ✓ Graphs
- ✓ Country highlights
- ✓ Labels
- ✓ Timeline graphics
- ✓ Globe renders

### All assets

- PNG
- Transparent
- 4K

---

# Image Prompt Rules

Every prompt should follow Vox editorial style.

### Requirements

- Editorial
- Documentary
- Clean
- Minimal
- Paper texture
- Topographic maps
- Data visualization
- Flat colors
- Soft shadows
- High quality
- No watermark
- No logos
- No text

---

# Camera Rules

Use cinematic movement.

### Allowed

- Slow Push
- Slow Zoom
- Truck Left
- Truck Right
- Tilt
- Orbit
- Parallax
- Map Zoom
- Satellite Zoom
- Split Screen
- Match Cut

---

# Animation Rules

Everything moves. Never show static images.

### Use

- Scale
- Rotation
- Position
- Opacity
- Mask reveal
- Stroke animation
- Line drawing
- Graph animation
- Timeline reveal
- Map reveal
- Counter animation
- Camera shake
- Paper movement
- Noise
- Film grain
- Stop motion

---

# Typography

Generate automatically.

### Styles

- Headline
- Subtitle
- Body
- Statistic
- Quote
- Label
- Timeline

### No random fonts. Use:

- Inter
- Helvetica
- Geist
- IBM Plex Sans

---

# Color Palette

| Token | Color |
|---|---|
| Primary Text | White |
| Secondary Text | Black |
| Muted | Dark Gray |
| Accent 1 | Vox Yellow |
| Accent 2 | Orange |
| Accent 3 | Blue |
| Alert | Red Accent |
| Background | Warm Paper |

---

# Music

Generate background music.

### Requirements

- Documentary
- Minimal
- Tension
- Ambient
- Cinematic
- No vocals
- Loopable

Generate: `music.wav`

---

# Sound Effects

### Generate

- Paper
- Swipe
- Whoosh
- Pop
- Typing
- Camera
- Impact
- Marker
- Click
- Page flip
- Wind
- Ocean
- Explosion (if needed)

Store: `/sfx`

---

# Voice

### Development: Edge TTS (Free)

Use `edge-tts` (Microsoft Edge Text-to-Speech) for development and testing.

```bash
# Install
pip install edge-tts

# Generate voice
edge-tts --voice "en-US-GuyNeural" --text "What if everything you knew about the ocean was wrong?" --write-media voice.wav
```

**Recommended Voices:**
- `en-US-GuyNeural` — Male, documentary narrator style
- `en-US-ChristopherNeural` — Male, calm authoritative
- `en-US-JennyNeural` — Female, clear professional
- `hi-IN-MadhurNeural` — Hindi male narrator

### Production: ElevenLabs

Switch to ElevenLabs API for final production renders.

### Requirements

- Natural
- Human
- Slow
- Emotion
- No robotic pauses

Export: `voice.wav`

---

# Timeline Builder

Automatically calculate:

- Scene Duration
- Voice Timing
- Animation Timing
- Music Timing
- Transition Timing

Export: `timeline.json`

---

# Remotion

Everything should be component driven.

### Components

- Background
- Map
- Image
- Headline
- Subtitle
- Graph
- Timeline
- Arrow
- Icon
- Statistics
- Camera
- Particles
- Noise
- Paper
- Transitions
- LowerThird
- SceneContainer

---

# Motion Rules

Every component accepts:

- startFrame
- endFrame
- duration
- delay
- scale
- rotation
- opacity
- position
- zIndex
- blur
- noise
- grain

---

# Transitions

### Allowed

- Push
- Slide
- Mask
- Paper Reveal
- Luma
- Blur
- Zoom
- Camera Cut
- Whip Pan
- Match Cut
- Graph Morph

Never use cheesy transitions.

---

# FFmpeg Final Polish

### Apply

- Film Grain
- Motion Blur
- Noise
- Color Grade
- Sharpen
- Limiter
- Loudness Normalize
- Fade
- Final Compression

### Export

- H264
- H265
- AV1
- 4K
- 60fps

---

# Folder Structure

```
project/
├── assets/
│   ├── images/
│   ├── maps/
│   ├── graphs/
│   ├── icons/
│   └── paper/
├── audio/
│   ├── music/
│   ├── voice/
│   └── sfx/
├── data/
│   ├── research.json
│   ├── script.json
│   ├── scene_plan.json
│   └── timeline.json
├── remotion/
│   ├── components/
│   ├── scenes/
│   ├── animations/
│   └── render/
└── final/
```

---

# Quality Checklist

- ✓ No empty frame
- ✓ No static visuals
- ✓ Motion every scene
- ✓ Voice synced
- ✓ Music synced
- ✓ Professional transitions
- ✓ Editorial layout
- ✓ Consistent colors
- ✓ Smooth camera
- ✓ High readability
- ✓ Vox quality

---

# Final Goal

Produce a premium documentary that looks professionally edited,
not AI-generated.

Everything should be generated automatically with minimal manual work.

Use Remotion for composition and animation.
Use FFmpeg for post-processing and encoding.

### Target quality:

- Netflix documentary
- Vox
- Johnny Harris
- MagnatesMedia
