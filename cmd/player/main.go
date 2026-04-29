package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourname/music-player/internal/engine"
	"github.com/yourname/music-player/internal/ui"
	"github.com/yourname/music-player/pkg/utils"
)

func main() {
	// ── 1. Resolve music directory ─────────────────────────────────────────
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// ── 2. Scan for audio files ────────────────────────────────────────────
	tracks, err := utils.ScanDir(absDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan error:", err)
		os.Exit(1)
	}

	// ── 3. Start engine ────────────────────────────────────────────────────
	eng, err := engine.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "engine error:", err)
		os.Exit(1)
	}

	// Load tracks into the queue.
	// engine.AddTrack calls playlist.AppendTrack (your linked-list method).
	for _, t := range tracks {
		eng.AddTrack(t)
	}

	// eng.Traverse()

	// ── 4. Launch Bubble Tea ───────────────────────────────────────────────
	prog := tea.NewProgram(
		ui.New(eng),
		tea.WithAltScreen(),
	)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ui error:", err)
		os.Exit(1)
	}
}
