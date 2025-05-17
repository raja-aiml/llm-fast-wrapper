package fiberapi

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/your-org/llm-fast-wrapper/internal/llm"
)

func Start() error {
	app := fiber.New()
	client := llm.NewOpenAIStreamer()

	app.Get("/stream", func(c *fiber.Ctx) error {
		prompt := c.Query("prompt", "hello")
		ch, err := client.Stream(prompt)
		if err != nil {
			return err
		}
		c.Set("Content-Type", "text/event-stream")
		return c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for token := range ch {
				msg := fmt.Sprintf("data: %s\n\n", token)
				if _, err := w.WriteString(msg); err != nil {
					log.Println("write error", err)
					return
				}
				w.Flush()
			}
		})
	})

	return app.Listen(":8080")
}
