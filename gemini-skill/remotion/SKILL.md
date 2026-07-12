---
name: gemini-remotion
description: >
  Use Remotion (React-based video composition framework) to build premium,
  high-end automated video templates, motion graphics, and documentaries.
  Use when the user wants to compose video elements programmatically, animate
  text overlays, maps, or charts, sync audio tracks, or automate rendering.
  Covers composition setup, spring physics, interpolation, and performance.
---

# 🎥 Remotion Video Composition & Motion Graphics Guide

This guide documents the API schemas, coding patterns, and rendering workflows for **Remotion**, a React-based programmatic video creation and animation framework.

---

## 🚀 Setup & Execution

### 1. Initialize a Project
Use `npx` to create a template project inside a directory (non-interactive mode):
```bash
npx -y create-video@latest ./remotion-app --template default
```

### 2. Local Preview Studio
Start the interactive developer interface to visual-debug frames, scrub timelines, and review props:
```bash
npm run start
```

### 3. Server/CLI Rendering
Render a specific composition to an MP4 file, optionally passing dynamic JSON properties:
```bash
npx remotion render src/index.ts [composition-id] out.mp4 --props=data/timeline.json
```

---

## 🎛️ Remotion Best Practices & Guidelines

### 1. Media Components
*   Always use Remotion’s optimized media elements instead of native HTML5 tags to ensure perfect frame-accurate synchronization:
    *   `import { Video, Audio, Img } from 'remotion';`
*   Use `<Sequence>` to shift or scope elements to a specific segment of the timeline:
    ```tsx
    import { Sequence } from 'remotion';
    
    // Renders the component starting at frame 30 for a duration of 120 frames
    <Sequence from={30} durationInFrames={120}>
      <MyTextComponent />
    </Sequence>
    ```

### 2. Deterministic Execution (No Randomness)
*   **Crucial Rule:** Remotion renders frame-by-frame. Never use standard non-deterministic calls like `Math.random()` or `new Date()`. This causes flickering and layout shifts during render.
*   **Solution:** Use Remotion’s deterministic helpers:
    ```tsx
    import { random } from 'remotion';
    
    const seededRandom = random("fixed-seed-string");
    ```

---

## 💫 Advanced Kinetic Animations (Vox/Johnny Harris Style)

To achieve premium, organic-feeling motion graphics, drive your styling parameters using physics-based **Springs** combined with **Interpolation**.

### 1. SNAPPY Spring Setup (Snappy reveal without excessive bounce)
```tsx
import { useCurrentFrame, useVideoConfig, spring, interpolate } from 'remotion';

export const KineticTitle: React.FC<{ text: string }> = ({ text }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Snappy spring configuration (damping: 18, stiffness: 120)
  const animDriver = spring({
    frame,
    fps,
    config: {
      damping: 18,
      stiffness: 120,
      mass: 0.8
    }
  });

  // Map the 0->1 spring driver to CSS transform styles
  const scale = interpolate(animDriver, [0, 1], [0.8, 1.0]);
  const translateY = interpolate(animDriver, [0, 1], [50, 0]);
  const opacity = interpolate(animDriver, [0, 0.8], [0, 1]);

  return (
    <h1 style={{
      transform: `translateY(${translateY}px) scale(${scale})`,
      opacity: opacity,
      fontFamily: 'IBM Plex Sans, sans-serif',
      fontSize: '70px',
      color: '#FFCC00',
      textAlign: 'center'
    }}>
      {text}
    </h1>
  );
};
```

### 2. Multi-Keyframe Interpolation (Fade In & Out)
Map multiple keyframes onto the frame counter to handle reveals and dismissals in a single declaration:
```tsx
const frame = useCurrentFrame();

// Fade-in during first 15 frames, stay visible, fade-out during final 15 frames (e.g., duration 150 frames)
const opacity = interpolate(
  frame,
  [0, 15, 135, 150], // Keyframe frames
  [0, 1, 1, 0],      // Output opacity levels
  { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' }
);
```

---

## 🎨 Data-Driven Video Pipeline (Chaining Scenes)

For AI automation, decouple the template code from the assets by reading from a central config (`timeline.json`).

### Root Registry (`src/Root.tsx`)
```tsx
import { Composition } from 'remotion';
import { MainVideo } from './MainVideo';
import timelineData from '../data/timeline.json';

export const Root: React.FC = () => {
  const fps = 24;
  const totalFrames = Math.floor(timelineData.total_duration_seconds * fps);

  return (
    <Composition
      id="VoxDoc"
      component={MainVideo}
      durationInFrames={totalFrames}
      fps={fps}
      width={1920}
      height={1080}
      defaultProps={{
        scenes: timelineData.scenes
      }}
    />
  );
};
```

### Main Compositor (`src/MainVideo.tsx`)
```tsx
import { Sequence, Audio } from 'remotion';
import { SceneContainer } from './components/SceneContainer';

interface Scene {
  scene_number: number;
  audio_file: string;
  audio_duration_seconds: number;
  image_file: string;
  visuals: any;
}

export const MainVideo: React.FC<{ scenes: Scene[] }> = ({ scenes }) => {
  let currentFrameOffset = 0;
  const fps = 24;

  return (
    <div style={{ backgroundColor: '#111', width: '100%', height: '100%' }}>
      {scenes.map((scene) => {
        const durationFrames = Math.floor(scene.audio_duration_seconds * fps);
        const startFrame = currentFrameOffset;
        currentFrameOffset += durationFrames;

        return (
          <Sequence
            key={scene.scene_number}
            from={startFrame}
            durationInFrames={durationFrames}
          >
            {/* Visual element with pan/zoom motion */}
            <SceneContainer scene={scene} durationFrames={durationFrames} />
            
            {/* Voiceover audio track synced directly */}
            <Audio src={scene.audio_file} />
          </Sequence>
        );
      })}
    </div>
  );
};
```
