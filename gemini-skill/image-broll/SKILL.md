---
name: gemini-image-broll
description: >
  Generate high-quality B-roll images using Google Imagen 3 via Gemini.
  Use when the user wants to create photorealistic photos, 3D renders,
  vector illustrations, flat lay compositions, or cinematic stills for
  thumbnails, banners, or video production. Supports image-to-image (I2I)
  with a reference image for style consistency.
---

# 🖼️ Gemini Image B-Roll Generation Guide (Imagen 3)

This guide documents the capabilities, prompting strategies, and integration options for generating high-quality B-roll images using Google Gemini's advanced image generation model (**Imagen 3**).

---

## 🚀 Overview & Specifications

Imagen 3 is Google's highest-quality text-to-image generator, producing detailed photorealistic pictures, vector art, 3D renders, and illustrations.

*   **Aspect Ratios:** Supports multiple ratios including **16:9 (Landscape)** for banners/desktop, **9:16 (Portrait)** for mobile, and **1:1 (Square)**.
*   **Image-to-Image (I2I):** You can supply a local reference image to guide the style, structure, and character consistency of the new output.
*   **Watermark Processing:** Generated images contain a transparent digital watermark that should be removed using the `logo-cleaner` tool before final use.

---

## 🎨 Creative Prompting Guide

To generate professional image B-roll, structure your prompt with these elements:
`[Subject/Focus] + [Composition & Framing] + [Lighting & Colors] + [Style & Details]`

### 1. Composition & Framing
*   **Flat Lay / Knolling:** Top-down view of items arranged neatly on a surface (great for tech/workspace setups).
*   **Shallow Depth of Field:** Blurred background that makes the foreground subject pop.
*   **Extreme Close-up:** Focuses on micro-textures and intricate details.
*   **Wide Angle:** Captures the entire environment or landscape.

### 2. Styles & Aesthetics
*   **Photorealistic:** High-fidelity, detailed textures, looking like real camera photographs.
*   **3D Render / Isometric:** Clean, modern digital art style (great for landing page illustrations).
*   **Minimalist Vector:** Flat colors, sharp lines, clean modern icons.

---

## 📋 Example Image B-Roll Prompt Templates

### 1. Developer Setup (Flat Lay / Knolling)
> **Prompt**: "A flat lay top-down view of a modern developer's desk. Includes a sleek dark laptop, mechanical keyboard, wireless mouse, headphones, and a clean notebook. Minimalist style, organized, dark wood background, soft diffused overhead lighting, photorealistic, 16:9 aspect ratio."

### 2. Cyberpunk Coding Space (Cinematic Close-up)
> **Prompt**: "A close-up shot of a modern programmer workspace at night. An ultra-wide curved monitor displaying neon blue code, reflection on the desk surface, glowing neon ambient strip lights. Cyberpunk aesthetic, moody volumetric lighting, photorealistic, 16:9 aspect ratio."

### 3. Abstract Tech / AI Illustration (3D Render)
> **Prompt**: "A modern 3D isometric render of an abstract brain network representing artificial intelligence. Glowing golden circuits, glassmorphic floating nodes, clean dark background. High-quality digital art, vibrant colors, 16:9 aspect ratio."

---

## 🔌 MCP Integration (`generate_image` Tool)

Invoke the image generation tool via the `gemini` MCP server:

```json
{
  "name": "generate_image",
  "arguments": {
    "prompt": "Your detailed image prompt here"
  }
}
```

### Optional Image-to-Image (I2I)
To guide generation using a reference image, pass the local file path:
```json
{
  "name": "generate_image",
  "arguments": {
    "prompt": "Change the lighting to warm sunset colors while keeping this style.",
    "ref_image_path": "/Users/akash/My-work/outputs/original_image.png"
  }
}
```

### Response Format
The tool returns the title, local container path, and download URL of the generated `.png` file:
```text
🎨 Image generated successfully!
URL: http://localhost:8002/output/img_r_xxxx_0.png
```
