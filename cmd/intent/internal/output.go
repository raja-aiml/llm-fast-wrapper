package internal

import (
	"fmt"

	"github.com/raja.aiml/llm-fast-wrapper/internal/intent"
)

func printDefaultStrategy() {
	fmt.Println("Matched:  Default Strategy (score=0.0000)")
	fmt.Printf("Strategy: built-in\n\n%s\n", intent.DefaultStrategy)
}

func printMatch(result *intent.MatchResult) {
	fmt.Printf("Matched:  %s (score=%.4f)\n", result.Name, result.Score)
	fmt.Printf("Strategy: %s\n\n%s\n", result.Path, result.Content)
}
