package ginapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raja.aiml/llm-fast-wrapper/internal/llm"
)

func Start() error {
		// set Gin to release mode to disable debug logs and warnings in production
		gin.SetMode(gin.ReleaseMode)
		// create a new Gin engine and attach Logger and Recovery middleware
		r := gin.New()
		r.Use(gin.Logger(), gin.Recovery())
		// disable trusting all proxies by default; configure as needed for your deployment
		if err := r.SetTrustedProxies(nil); err != nil {
			return err
		}
	client := llm.NewOpenAIStreamer()

	r.POST("/v1/chat/completions", func(c *gin.Context) {
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Stream bool `json:"stream"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !req.Stream {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stream must be true"})
			return
		}

		var prompt string
		for _, m := range req.Messages {
			if m.Role == "user" {
				prompt += m.Content + " "
			}
		}

		ch, err := client.Stream(prompt)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()
		enc := json.NewEncoder(c.Writer)
		for chunk := range ch {
			if _, err := c.Writer.Write([]byte("data: ")); err != nil {
				log.Println("write error", err)
				return
			}
			if err := enc.Encode(chunk); err != nil {
				log.Println("encode error", err)
				return
			}
			if _, err := c.Writer.Write([]byte("\n")); err != nil {
				log.Println("write error", err)
				return
			}
			c.Writer.Write([]byte("\n"))
			c.Writer.Flush()
		}
		c.Writer.Write([]byte("data: [DONE]\n\n"))
		c.Writer.Flush()
	})

	return r.Run(":8080")
}
