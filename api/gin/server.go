package ginapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func Start() error {
	r := gin.Default()
	client := llm.NewOpenAIStreamer()

	r.GET("/stream", func(c *gin.Context) {
		prompt := c.Query("prompt")
		ch, err := client.Stream(prompt)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()
		for token := range ch {
			msg := fmt.Sprintf("data: %s\n\n", token)
			if _, err := c.Writer.Write([]byte(msg)); err != nil {
				log.Println("write error", err)
				return
			}
			c.Writer.Flush()
		}
	})

	return r.Run(":8080")
}
