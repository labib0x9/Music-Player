// Package ui contains the Bubble Tea TUI.
//
// Design principle: the UI is a pure renderer of engine.Snapshot.
// It NEVER calls playlist methods directly — that is the engine's job.
// The flow is:
//
//	keypress → tea.Cmd → engine method → next tick snapshot → re-render
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourname/music-player/internal/engine"
)

// ─── Messages ─────────────────────────────────────────────────────────────────

// tickMsg is sent every 500 ms to refresh elapsed time.
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ─── Styles ───────────────────────────────────────────────────────────────────

var (
	colorGreen  = lipgloss.Color("#00FF87")
	colorYellow = lipgloss.Color("#FFD700")
	colorGray   = lipgloss.Color("#555555")
	colorWhite  = lipgloss.Color("#EEEEEE")
	colorDim    = lipgloss.Color("#888888")
	colorBg     = lipgloss.Color("#111111")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen).
			Padding(0, 1)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("#1A3A2A")).
			Foreground(colorGreen).
			Bold(true)

	styleNormal = lipgloss.NewStyle().
			Foreground(colorWhite)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleStatusBar = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorGreen).
			Padding(0, 1)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1E3E2E")).
			Padding(0, 1)
)

// ─── Model ────────────────────────────────────────────────────────────────────

// Model is the Bubble Tea model. It holds a reference to the engine and the
// last rendered snapshot.
type Model struct {
	eng      *engine.Engine
	snap     engine.Snapshot
	cursor   int       // UI cursor (may differ from playing index)
	progress progress.Model
	width    int
	height   int
}

// New returns an initialised UI Model.
func New(eng *engine.Engine) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	m := Model{
		eng:      eng,
		progress: p,
	}
	m.snap = eng.State()
	return m
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tick()
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		return m, nil

	case tickMsg:
		m.snap = m.eng.State()
		return m, tick()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tracks := m.snap.Tracks

	switch msg.String() {

	// ── Quit ──────────────────────────────────────────────────────────
	case "q", "ctrl+c":
		m.eng.Stop()
		return m, tea.Quit

	// ── Cursor navigation ─────────────────────────────────────────────
	case "j", "down":
		if m.cursor < len(tracks)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "g":
		m.cursor = 0

	case "G":
		if len(tracks) > 0 {
			m.cursor = len(tracks) - 1
		}

	// ── Playback ──────────────────────────────────────────────────────
	case "enter":
		if len(tracks) > 0 {
			m.eng.JumpTo(tracks[m.cursor].ID)
		}

	case " ":
		m.eng.Pause()

	case "s":
		m.eng.Stop()

	// ── Track navigation ──────────────────────────────────────────────
	case "n":
		m.eng.Next()
		m.sync()

	case "p":
		m.eng.Prev()
		m.sync()

	// ── Repeat ────────────────────────────────────────────────────────
	case "r":
		m.eng.ToggleRepeat()
	}

	m.snap = m.eng.State()
	return m, nil
}

// sync moves the UI cursor to the currently playing track.
func (m *Model) sync() {
	m.snap = m.eng.State()
	m.cursor = m.snap.CursorIndex
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")
	sb.WriteString(m.renderTrackList())
	sb.WriteString("\n")
	sb.WriteString(m.renderNowPlaying())
	sb.WriteString("\n")
	sb.WriteString(m.renderHelp())

	return sb.String()
}

func (m Model) renderHeader() string {
	return styleTitle.Render("♪  music-player")
}

func (m Model) renderTrackList() string {
	tracks := m.snap.Tracks
	if len(tracks) == 0 {
		return styleDim.Render("  (no tracks loaded — pass a directory as argument)")
	}

	// Show at most (height - 10) rows so the now-playing bar always fits.
	maxRows := m.height - 10
	if maxRows < 5 {
		maxRows = 5
	}

	// Windowing: keep cursor visible.
	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(tracks) {
		end = len(tracks)
	}

	var rows []string
	for i := start; i < end; i++ {
		t := tracks[i]
		line := fmt.Sprintf(" %3d.  %-35s  %s", t.ID, truncate(t.Title, 35), styleDim.Render(t.Artist))

		isPlaying := m.snap.CurrentTrack != nil && m.snap.CurrentTrack.ID == t.ID
		if isPlaying {
			indicator := lipgloss.NewStyle().Foreground(colorGreen).Render("▶ ")
			line = indicator + strings.TrimLeft(line, " ")
		}

		if i == m.cursor {
			rows = append(rows, styleSelected.Render(line))
		} else {
			rows = append(rows, styleNormal.Render(line))
		}
	}

	return styleBorder.Render(strings.Join(rows, "\n"))
}

func (m Model) renderNowPlaying() string {
	snap := m.snap

	if snap.CurrentTrack == nil {
		return styleStatusBar.Render("  ■  stopped")
	}

	icon := "▶"
	if snap.State == engine.StatePaused {
		icon = "⏸"
	}

	title := truncate(snap.CurrentTrack.Title, 40)
	artist := snap.CurrentTrack.Artist

	elapsed := fmtDuration(snap.Elapsed)
	total := fmtDuration(snap.Total)

	var pct float64
	if snap.Total > 0 {
		pct = float64(snap.Elapsed) / float64(snap.Total)
		if pct > 1 {
			pct = 1
		}
	}
	progressBar := m.progress.ViewAs(pct)

	repeatTag := ""
	if snap.RepeatOne {
		repeatTag = lipgloss.NewStyle().Foreground(colorYellow).Render("  ↺")
	}

	line1 := fmt.Sprintf("  %s  %s — %s%s", icon, title, styleDim.Render(artist), repeatTag)
	line2 := fmt.Sprintf("  %s  %s / %s", progressBar, elapsed, total)

	return styleBorder.Render(line1 + "\n" + line2)
}

func (m Model) renderHelp() string {
	keys := []string{
		"↑/k up", "↓/j down", "enter play", "space pause",
		"n next", "p prev", "r repeat", "s stop", "q quit",
	}
	return styleHelp.Render("  " + strings.Join(keys, "  ·  "))
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
