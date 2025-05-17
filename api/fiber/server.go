package fiberapi

import (
	"bufio"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func Start() error {
	app := fiber.New()

	client := llm.NewOpenAIStreamer()

	app.Get("/stream", func(c *fiber.Ctx) error {
		prompt := c.Query("prompt", "hello")

		ch, err := client.Stream(prompt)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		c.Set("Content-Type", "text/event-stream")
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for token := range ch {
				msg := fmt.Sprintf("data: %s\n\n", token)
				if _, err := w.WriteString(msg); err != nil {
					log.Println("write error:", err)
					return
				}
				w.Flush()
			}
		})

		return nil
	})

	log.Println("[INFO] Fiber server listening on :8080")
	return app.Listen(":8080")
}
