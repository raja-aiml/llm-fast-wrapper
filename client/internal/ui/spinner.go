package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Spinner struct {
	style lipgloss.Style
	done  chan struct{}
}

func NewSpinner() *Spinner {
	return &Spinner{
		style: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34")),
		done:  make(chan struct{}),
	}
}

// Start displays a spinning "Thinking..." animation.
func (s *Spinner) Start() {
	go func() {
		frames := []rune{'|', '/', '-', '\\'}
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("\r%s %c", s.style.Render("Thinking..."), frames[i%len(frames)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner and clears the line.
func (s *Spinner) Stop() {
	close(s.done)
	fmt.Printf("\r")
}
