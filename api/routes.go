package api

import (
	"goapi/gemini"

	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all HTTP API routes onto the Fiber app instance
func RegisterRoutes(app *fiber.App) {
	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":                   "online",
			"engine":                   "needle2 + free-gemini-api (pure go)",
			"active_extension_workers": gemini.GetActiveWorkerCount(),
			"cookie_accounts_in_pool":  gemini.GetActiveAccountCount(),
		})
	})

	// OpenAI-compatible models list
	app.Get("/v1/models", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"object": "list",
			"data": []fiber.Map{
				{"id": "gemini-3.7-flash", "object": "model", "owned_by": "free-gemini-api"},
			},
		})
	})

	// OpenAI-compatible chat completions (with Native Needle 2 Tool Calling)
	app.Post("/v1/chat/completions", HandleOpenAIChatCompletions)

	// Unified chat endpoint (text, images, screen recordings, ref videos)
	app.Post("/chat", HandleUnifiedChat)

	// Gemini Music generation
	app.Post("/music", HandleMusic)

	// Reset active session
	app.Post("/reset", func(c fiber.Ctx) error {
		type ResetRequest struct {
			UserID string `json:"user_id"`
		}
		var req ResetRequest
		c.Bind().JSON(&req)

		sessionID := req.UserID
		if sessionID == "" {
			sessionID = c.IP()
		}

		if _, ok := UserSessions.Load(sessionID); ok {
			UserSessions.Delete(sessionID)
			return c.JSON(fiber.Map{"status": "success", "message": "Session reset"})
		}

		return c.JSON(fiber.Map{"status": "error", "message": "No active session"})
	})

	// Session status & health info
	app.Get("/status", func(c fiber.Ctx) error {
		sessionID := c.Query("user_id")
		if sessionID == "" {
			sessionID = c.IP()
		}

		if clientRaw, ok := UserSessions.Load(sessionID); ok {
			client := clientRaw.(*gemini.GeminiClient)
			return c.JSON(fiber.Map{
				"session_id":      sessionID,
				"initialized":     client.IsInitialized,
				"conversation_id": client.ConversationID,
				"expired":         false,
			})
		}
		return c.JSON(fiber.Map{"session_id": sessionID, "active": false})
	})

	// Static generated output files (images, videos, music)
	app.Get("/output/*", func(c fiber.Ctx) error {
		return c.SendFile("./output/" + c.Params("*"))
	})
}
