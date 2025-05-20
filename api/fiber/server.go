package fiberapi

import (
	"bufio"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func Start() error {
	app := fiber.New()

	client := llm.NewOpenAIStreamer()

	app.Post("/v1/chat/completions", func(c *fiber.Ctx) error {
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Stream bool `json:"stream"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if !req.Stream {
			return fiber.NewError(fiber.StatusBadRequest, "stream must be true")
		}

		var prompt string
		for _, m := range req.Messages {
			if m.Role == "user" {
				prompt += m.Content + " "
			}
		}

		ch, err := client.Stream(c.Context(), prompt)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		c.Set("Content-Type", "text/event-stream")
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			enc := json.NewEncoder(w)
			for chunk := range ch {
				w.WriteString("data: ")
				if err := enc.Encode(chunk); err != nil {
					log.Println("encode error:", err)
					return
				}
				w.WriteString("\n\n")
				w.Flush()
			}
			w.WriteString("data: [DONE]\n\n")
			w.Flush()
		})

		return nil
	})

	log.Println("[INFO] Fiber server listening on :8080")
	return app.Listen(":8080")
}
