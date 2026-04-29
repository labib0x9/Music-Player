// Package audio provides a thin wrapper around faiface/beep for
// play / pause / resume / stop operations.
// It runs the playback loop in its own goroutine and communicates
// progress and end-of-track events via channels.
package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

const sampleRate = beep.SampleRate(44100)

// DoneMsg is sent on the Done channel when a track finishes naturally.
type DoneMsg struct{}

// Backend manages a single audio stream.
type Backend struct {
	mu      sync.Mutex
	ctrl    *beep.Ctrl   // pause / resume
	done    chan DoneMsg  // signals end-of-track to engine
	closeFn func() error // closes the current decoder
	elapsed time.Duration
	total   time.Duration
	started bool
}

// New initialises the speaker once and returns a ready Backend.
func New() (*Backend, error) {
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		return nil, fmt.Errorf("audio: speaker init: %w", err)
	}
	return &Backend{
		done: make(chan DoneMsg, 1),
	}, nil
}

// Done returns the channel that fires when the current track ends.
func (b *Backend) Done() <-chan DoneMsg { return b.done }

// Load opens a file and prepares it for playback (does not start playing).
func (b *Backend) Load(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stop() // stop any existing playback

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("audio: open %q: %w", path, err)
	}

	var (
		streamer beep.StreamSeekCloser
		format   beep.Format
	)

	switch filepath.Ext(path) {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
	default:
		f.Close()
		return fmt.Errorf("audio: unsupported format %q", filepath.Ext(path))
	}
	if err != nil {
		f.Close()
		return fmt.Errorf("audio: decode %q: %w", path, err)
	}

	// Resample if the file sample rate differs from speaker rate.
	var stream beep.Streamer = streamer
	if format.SampleRate != sampleRate {
		stream = beep.Resample(4, format.SampleRate, sampleRate, streamer)
	}

	b.ctrl = &beep.Ctrl{Streamer: beep.Seq(stream, beep.Callback(func() {
		b.done <- DoneMsg{}
	}))}
	b.closeFn = func() error { return streamer.Close() }

	samples := streamer.Len()
	b.total = format.SampleRate.D(samples)
	b.elapsed = 0
	b.started = false

	return nil
}

// Play starts or resumes playback.
func (b *Backend) Play() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ctrl == nil {
		return
	}
	if !b.started {
		speaker.Play(b.ctrl)
		b.started = true
		return
	}
	speaker.Lock()
	b.ctrl.Paused = false
	speaker.Unlock()
}

// Pause pauses playback without unloading the stream.
func (b *Backend) Pause() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ctrl == nil {
		return
	}
	speaker.Lock()
	b.ctrl.Paused = true
	speaker.Unlock()
}

// Stop halts playback and closes the current stream.
func (b *Backend) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stop()
}

// stop is the internal (already-locked) variant.
func (b *Backend) stop() {
	if b.ctrl == nil {
		return
	}
	speaker.Clear()
	if b.closeFn != nil {
		_ = b.closeFn()
		b.closeFn = nil
	}
	b.ctrl = nil
	b.started = false
}

// IsPaused reports whether the stream is paused.
func (b *Backend) IsPaused() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ctrl == nil {
		return false
	}
	speaker.Lock()
	p := b.ctrl.Paused
	speaker.Unlock()
	return p
}

// Total returns the duration of the loaded track.
func (b *Backend) Total() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.total
}
