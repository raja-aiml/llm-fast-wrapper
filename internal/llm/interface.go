package llm

type Streamer interface {
	Stream(prompt string) (<-chan string, error)
}
