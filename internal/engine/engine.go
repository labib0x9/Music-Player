// Package engine is the heart of the player.
// It owns the Playlist and the Audio Backend and exposes a command-based API
// that the UI calls.  All state transitions happen here — the UI never touches
// the audio backend directly.
//
// Architecture:
//
//	UI  ──cmd──►  Engine  ──calls──►  playlist.Playlist
//	                │                      (YOUR linked list)
//	                └──calls──►  audio.Backend
//	UI  ◄──state── Engine (polled via State())
package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourname/music-player/internal/audio"
	"github.com/yourname/music-player/internal/model"
	"github.com/yourname/music-player/internal/playlist"
)

// PlayerState represents what the engine is currently doing.
type PlayerState int

const (
	StateStopped PlayerState = iota
	StatePlaying
	StatePaused
)

func (s PlayerState) String() string {
	switch s {
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	default:
		return "stopped"
	}
}

// Snapshot is a read-only view of engine state that the UI renders.
// The UI should only ever read this — never mutate engine internals.
type Snapshot struct {
	State       PlayerState
	CurrentTrack *model.Track  // nil when stopped/empty
	Tracks      []model.Track  // ordered snapshot of the queue
	Elapsed     time.Duration
	Total       time.Duration
	CursorIndex int            // which track in Tracks is currently playing
	RepeatOne   bool
}

// Engine wires together the playlist (linked list) and audio backend.
type Engine struct {
	mu       sync.RWMutex
	pl       *playlist.Playlist
	audio    *audio.Backend
	state    PlayerState
	elapsed  time.Duration
	start    time.Time  // time.Time when last Play() was called
	repeatOne bool
}

// New creates a ready Engine. Call Close() when done.
func New() (*Engine, error) {
	ab, err := audio.New()
	if err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	e := &Engine{
		pl:    playlist.New(),
		audio: ab,
		state: StateStopped,
	}
	go e.watchDone()
	return e, nil
}

// ─── Queue management ─────────────────────────────────────────────────────────

// AddTrack appends a track to the queue.
//
// Calls: playlist.AppendTrack  ← YOUR linked-list method
func (e *Engine) AddTrack(t model.Track) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pl.AppendTrack(t)
}

// RemoveCurrent removes the current track and plays the next one if running.
//
// Calls: playlist.RemoveCurrent  ← YOUR linked-list method
func (e *Engine) RemoveCurrent() {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.pl.RemoveCurrent()
	if !ok {
		return
	}
	// If we were playing, continue with whatever is now current.
	if e.state == StatePlaying {
		e.loadAndPlay()
	}
}

// ─── Playback commands ────────────────────────────────────────────────────────

// Play starts or resumes playback of the current track.
func (e *Engine) Play() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.pl.IsEmpty() {
		return
	}
	switch e.state {
	case StatePaused:
		e.audio.Play()
		e.start = time.Now()
		e.state = StatePlaying
	case StateStopped:
		e.loadAndPlay()
	}
}

// Pause toggles pause/resume.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch e.state {
	case StatePlaying:
		e.audio.Pause()
		e.elapsed += time.Since(e.start)
		e.state = StatePaused
	case StatePaused:
		e.audio.Play()
		e.start = time.Now()
		e.state = StatePlaying
	}
}

// Stop halts playback.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.audio.Stop()
	e.state = StateStopped
	e.elapsed = 0
}

// Next skips to the next track.
//
// Calls: playlist.Next  ← YOUR linked-list method
func (e *Engine) Next() {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.pl.Next()
	if !ok {
		// Already at end; stop.
		e.audio.Stop()
		e.state = StateStopped
		return
	}
	if e.state == StatePlaying || e.state == StatePaused {
		e.loadAndPlay()
	}
}

// Prev goes back to the previous track.
//
// Calls: playlist.Prev  ← YOUR linked-list method
func (e *Engine) Prev() {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.pl.Prev()
	if !ok {
		return
	}
	if e.state == StatePlaying || e.state == StatePaused {
		e.loadAndPlay()
	}
}

// JumpTo plays the track with the given ID.
//
// Calls: playlist.JumpTo  ← YOUR linked-list method
func (e *Engine) JumpTo(id int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, ok := e.pl.JumpTo(id)
	if !ok {
		return
	}
	e.loadAndPlay()
}

// ToggleRepeat flips the repeat-one mode.
func (e *Engine) ToggleRepeat() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.repeatOne = !e.repeatOne
}

// ─── State snapshot ───────────────────────────────────────────────────────────

// State returns a consistent read-only snapshot for the UI to render.
func (e *Engine) State() Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	snap := Snapshot{
		State:     e.state,
		RepeatOne: e.repeatOne,
		Total:     e.audio.Total(),
	}

	// Elapsed time: committed elapsed + time since last Play().
	if e.state == StatePlaying {
		snap.Elapsed = e.elapsed + time.Since(e.start)
	} else {
		snap.Elapsed = e.elapsed
	}
	if snap.Elapsed > snap.Total && snap.Total > 0 {
		snap.Elapsed = snap.Total
	}

	// ── Calls: playlist.ToSlice / playlist.Current ────────────────────
	// (YOUR linked-list methods)
	snap.Tracks = e.pl.ToSlice()

	curr := e.pl.Current()
	if curr != nil {
		t := curr.Track
		snap.CurrentTrack = &t
		// Find cursor index in the flat slice for UI highlight.
		for i, tr := range snap.Tracks {
			if tr.ID == t.ID {
				snap.CursorIndex = i
				break
			}
		}
	}

	return snap
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// loadAndPlay loads the current node's file and begins playback.
// Must be called with e.mu held (write lock).
func (e *Engine) loadAndPlay() {
	curr := e.pl.Current()
	if curr == nil {
		return
	}
	e.audio.Stop()
	if err := e.audio.Load(curr.Track.Path); err != nil {
		// Skip broken files: advance and try again.
		if _, ok := e.pl.Next(); ok {
			e.loadAndPlay()
		} else {
			e.state = StateStopped
		}
		return
	}
	e.elapsed = 0
	e.start = time.Now()
	e.audio.Play()
	e.state = StatePlaying
}

// watchDone listens for end-of-track signals from the audio backend
// and auto-advances to the next track (or repeats).
func (e *Engine) watchDone() {
	for range e.audio.Done() {
		e.mu.Lock()
		if e.repeatOne {
			// Replay same track.
			e.loadAndPlay()
		} else {
			_, ok := e.pl.Next() // YOUR linked-list method
			if ok {
				e.loadAndPlay()
			} else {
				e.state = StateStopped
				e.elapsed = 0
			}
		}
		e.mu.Unlock()
	}
}
