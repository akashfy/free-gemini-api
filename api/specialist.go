package api

import (
	"encoding/json"
	"fmt"
	"goapi/gemini"
	"log"
	"strings"
)

// GeminiSpecialistTools defines the core specialized functions of this API
var GeminiSpecialistTools = []map[string]any{
	{
		"name":        "generate_image",
		"description": "Generate high-resolution 8K photos, illustrations, art, or UI mockups using Gemini Image",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "Detailed visual description of the image to generate",
				},
				"aspect_ratio": map[string]any{
					"type":        "string",
					"enum":        []string{"1:1", "16:9", "9:16", "4:3", "3:4"},
					"description": "Aspect ratio of the generated image",
				},
			},
			"required": []string{"prompt"},
		},
	},
	{
		"name":        "generate_video",
		"description": "Generate cinematic HD video clips with motion, lighting, and camera movements using Gemini Video",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "Detailed description of the scene and video motion",
				},
			},
			"required": []string{"prompt"},
		},
	},
	{
		"name":        "generate_music",
		"description": "Generate songs, background scores, lofi beats, or audio tracks using Gemini Music",
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "Description of the music track, genre, tempo, instruments, and mood",
				},
			},
			"required": []string{"prompt"},
		},
	},
}

type TriageResult struct {
	Intent         string // "chat", "image", "video", "music", "tool_call"
	EnhancedPrompt string
	Arguments      map[string]any
}

// TriagePrompt analyzes user intent in <15ms using Needle 2 engine and routes to the best Gemini engine
func TriagePrompt(prompt string, userTools []map[string]any) (*TriageResult, error) {
	engine, err := gemini.GetNeedleEngine()
	if err != nil {
		return &TriageResult{Intent: "chat", EnhancedPrompt: prompt}, nil
	}

	// Merge specialist tools with any user-provided tools
	allTools := append([]map[string]any{}, GeminiSpecialistTools...)
	if len(userTools) > 0 {
		for _, ut := range userTools {
			if ut["type"] == "function" && ut["function"] != nil {
				if fn, ok := ut["function"].(map[string]any); ok {
					allTools = append(allTools, map[string]any{
						"name":        fn["name"],
						"description": fn["description"],
						"parameters":  fn["parameters"],
					})
				}
			} else {
				allTools = append(allTools, ut)
			}
		}
	}

	toolsJSONBytes, _ := json.Marshal(allTools)
	needleResp, cErr := engine.CompleteTools(prompt, string(toolsJSONBytes))
	if cErr != nil || needleResp == nil {
		return &TriageResult{Intent: "chat", EnhancedPrompt: prompt}, nil
	}

	if needleResp.Type == "call" && len(needleResp.FunctionCalls) > 0 {
		fc := needleResp.FunctionCalls[0]
		switch fc.Name {
		case "generate_image":
			p, _ := fc.Arguments["prompt"].(string)
			if p == "" {
				p = prompt
			}
			log.Printf("🎯 [Needle Specialist] Triaged as Image Generation: %s", p)
			return &TriageResult{
				Intent:         "image",
				EnhancedPrompt: fmt.Sprintf("Generate an image of %s", strings.TrimPrefix(p, "Generate an image of ")),
				Arguments:      fc.Arguments,
			}, nil

		case "generate_video":
			p, _ := fc.Arguments["prompt"].(string)
			if p == "" {
				p = prompt
			}
			log.Printf("🎯 [Needle Specialist] Triaged as Video Generation: %s", p)
			return &TriageResult{
				Intent:         "video",
				EnhancedPrompt: fmt.Sprintf("Generate a video of %s", strings.TrimPrefix(p, "Generate a video of ")),
				Arguments:      fc.Arguments,
			}, nil

		case "generate_music":
			p, _ := fc.Arguments["prompt"].(string)
			if p == "" {
				p = prompt
			}
			log.Printf("🎯 [Needle Specialist] Triaged as Music Generation: %s", p)
			return &TriageResult{
				Intent:         "music",
				EnhancedPrompt: p,
				Arguments:      fc.Arguments,
			}, nil

		default:
			// Custom user tool
			log.Printf("🎯 [Needle Specialist] Triaged as Custom Tool Call [%s]", fc.Name)
			return &TriageResult{
				Intent:    "tool_call",
				Arguments: fc.Arguments,
			}, nil
		}
	}

	return &TriageResult{Intent: "chat", EnhancedPrompt: prompt}, nil
}
