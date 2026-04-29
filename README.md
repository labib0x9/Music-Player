# music-player

A terminal music player built in Go using Bubble Tea for TUI and a **doubly linked list** as the playback queue. 

**Claude generated to understand linked-list**

## Architecture

```
cmd/player/main.go
        │
        ▼
internal/engine/engine.go   ← state machine + concurrency
        │              │
        ▼              ▼
internal/playlist/   internal/audio/
  playlist.go           backend.go
  (YOUR TASK)           (beep wrapper)
        │
        ▼
internal/model/track.go   ← Track + Node types
        
internal/ui/model.go      ← Bubble Tea, reads engine.Snapshot
pkg/utils/scanner.go      ← directory walker
```

**Data flow:**

```
keypress → ui.handleKey → eng.method → playlist.method (your code)
                                     ↘ audio.Backend
tick (500ms) → eng.State() → Snapshot → ui.View()
```

## Build & Run

```bash
go mod tidy
go run ./cmd/player /path/to/music
```

Supported formats: `.mp3`, `.wav`, `.flac`, `.ogg`

## Key Bindings

| Key       | Action              |
|-----------|---------------------|
| `↑` / `k` | move cursor up      |
| `↓` / `j` | move cursor down    |
| `enter`   | play selected track |
| `space`   | pause / resume      |
| `n`       | next track          |
| `p`       | previous track      |
| `r`       | toggle repeat-one   |
| `s`       | stop                |
| `q`       | quit                |

---

## Your Task — Implement the Linked List

All the logic in **`internal/playlist/playlist.go`** is left for you.

### Checklist

- [done] `AppendTrack(t Track)`  
  → insert at tail, update `Head`/`Tail`/`Curr`, increment `Size`

- [done] `PrependTrack(t Track)`  
  → insert at head, update `Head`/`Tail`/`Curr`, increment `Size`

- [ ] `RemoveCurrent() (Track, bool)`  
  → unlink `Curr`, re-link neighbours, update `Head`/`Tail`, decrement `Size`  
  → advance `Curr` to Next (or Prev if at tail)

- [done] `Next() (*Node, bool)`  
  → move `Curr = Curr.Next`, return it; return `nil, false` at tail

- [done] `Prev() (*Node, bool)`  
  → move `Curr = Curr.Prev`, return it; return `nil, false` at head

- [ ] `JumpTo(id int) (*Node, bool)`  
  → linear scan, set `Curr` to matching node

- [ ] `ToSlice() []Track`  
  → walk head→tail, collect all `Track` values

### Edge Cases to Handle

| Case | Affected methods |
|------|-----------------|
| Empty list | all |
| Single node | `RemoveCurrent`, `Next`, `Prev` |
| Remove head | `RemoveCurrent` |
| Remove tail | `RemoveCurrent` |
| ID not found | `JumpTo` |

### Complexity Targets

| Method | Target |
|--------|--------|
| `AppendTrack` | O(1) |
| `PrependTrack` | O(1) |
| `Next` / `Prev` | O(1) |
| `RemoveCurrent` | O(1) |
| `JumpTo` | O(n) — acceptable |
| `ToSlice` | O(n) — acceptable |

---

## What Is Already Done (Don't Touch)

| File | Status |
|------|--------|
| `internal/model/track.go` | ✅ complete |
| `internal/audio/backend.go` | ✅ complete |
| `internal/engine/engine.go` | ✅ complete |
| `internal/ui/model.go` | ✅ complete |
| `pkg/utils/scanner.go` | ✅ complete |
| `cmd/player/main.go` | ✅ complete |
| `internal/playlist/playlist.go` | ⛔ **your task** |
