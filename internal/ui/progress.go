package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// ProgressBar represents a progress bar for batch operations
type ProgressBar struct {
	total     int
	current   int
	startTime time.Time
	width     int
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total: total,
		width: 50, // Default width of the progress bar
	}
}

// Start starts the progress bar
func (p *ProgressBar) Start() {
	p.startTime = time.Now()
	p.current = 0

	// Get terminal height
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = 24 // Default height if we can't get terminal size
	}

	// Move cursor to bottom line and save position
	fmt.Printf("\033[%d;0H", height) // Move to bottom line
	fmt.Print("\033[s")              // Save cursor position
	p.update()
}

// Stop stops the progress bar
func (p *ProgressBar) Stop() {
	p.current = p.total
	p.update()
	// Move cursor to next line after progress bar
	fmt.Println()
}

// Increment increments the progress bar
func (p *ProgressBar) Increment() {
	p.current++
	// Restore cursor position before updating
	fmt.Print("\033[u") // Restore cursor position
	p.update()
}

// update updates the progress bar display
func (p *ProgressBar) update() {
	// Calculate progress percentage
	percent := float64(p.current) / float64(p.total) * 100

	// Calculate progress bar width
	completed := int(float64(p.width) * float64(p.current) / float64(p.total))
	remaining := p.width - completed

	// Calculate elapsed time and estimated time remaining
	elapsed := time.Since(p.startTime)
	var remainingTime time.Duration
	if p.current > 0 {
		remainingTime = time.Duration(float64(elapsed) * float64(p.total-p.current) / float64(p.current))
	}

	// Calculate speed (items per second)
	speed := float64(p.current) / elapsed.Seconds()

	// Create progress bar string
	bar := strings.Repeat("=", completed) + strings.Repeat("-", remaining)

	// Clear the line and print progress bar
	fmt.Print("\033[2K") // Clear the line
	fmt.Printf("\r[%s] %.1f%% (%d/%d) %.1f items/sec ETA: %s",
		bar,
		percent,
		p.current,
		p.total,
		speed,
		remainingTime.Round(time.Second),
	)
}
