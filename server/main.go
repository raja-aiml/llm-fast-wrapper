package main

import (
	"log"

	"github.com/raja.aiml/llm-fast-wrapper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("cli error: %v", err)
	}
}
