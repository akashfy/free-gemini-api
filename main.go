package main

import (
	"goapi/api"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env", "config.env"); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Check if invoked as MCP stdio server
	if len(os.Args) > 1 && (os.Args[1] == "--mcp" || os.Args[1] == "-mcp" || os.Args[1] == "mcp") {
		api.RunMCPServer()
		return
	}

	// Start Chrome Extension WebSocket bridge
	api.StartWebSocketBridge()

	// Initialize Fiber web application
	app := fiber.New(fiber.Config{
		AppName:   "Gemini Go API (Gemini 3.7 + Needle 2)",
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ RECOVERED from panic: %v", r)
				c.Status(500).JSON(fiber.Map{"error": "Internal Server Error - Recovered"})
			}
		}()
		return c.Next()
	})

	// Register all HTTP routes
	api.RegisterRoutes(app)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Printf("🚀 Free Gemini API Server starting on port :%s...", port)
	log.Fatal(app.Listen(":" + port))
}
