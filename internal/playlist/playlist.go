// Package playlist manages the active playback queue using a doubly linked list.
//
// ╔══════════════════════════════════════════════════════════════╗
// ║  YOUR TASK: implement every function marked with TODO below. ║
// ║  The engine and UI depend on the Playlist type — the         ║
// ║  function signatures must stay exactly as written.           ║
// ╚══════════════════════════════════════════════════════════════╝
package playlist

import (
	"fmt"

	"github.com/yourname/music-player/internal/model"
)

// Playlist is a doubly linked list used as the active playback queue.
type Playlist struct {
	Head *model.Node
	Tail *model.Node
	Curr *model.Node
	Size int
}

// New returns an empty Playlist.
func New() *Playlist {
	return &Playlist{}
}

// ─── Insertion ───────────────────────────────────────────────────────────────

// AppendTrack adds a track to the tail of the list.
//
// TODO: implement
//   - create a new Node wrapping t
//   - if the list is empty, set Head, Tail, and Curr to the new node
//   - otherwise link the new node after Tail and update Tail
//   - increment Size
//   - O(1)
func (p *Playlist) AppendTrack(t model.Track) {
	// panic("AppendTrack: not implemented")
	node := &model.Node{
		Track: t,
		Prev:  p.Tail,
		Next:  nil,
	}

	if p.Size == 0 {
		p.Curr = node
		p.Head = node
		p.Tail = node
	} else {
		p.Tail.Next = node
		p.Tail = p.Tail.Next
	}

	p.Size++
}

// PrependTrack adds a track to the head of the list.
//
// TODO: implement
//   - create a new Node wrapping t
//   - if the list is empty, set Head, Tail, and Curr to the new node
//   - otherwise link the new node before Head and update Head
//   - increment Size
//   - O(1)
func (p *Playlist) PrependTrack(t model.Track) {
	// panic("PrependTrack: not implemented")
}

// ─── Deletion ─────────────────────────────────────────────────────────────────

// RemoveCurrent removes Curr from the list and advances Curr to Next (or Prev
// if Curr was Tail). Returns the removed track.
//
// TODO: implement
//   - handle the edge cases: list is empty, single-node list
//   - re-link Prev and Next neighbours
//   - update Head/Tail if needed
//   - decrement Size
//   - return the removed track
//   - O(1)
func (p *Playlist) RemoveCurrent() (model.Track, bool) {
	// panic("RemoveCurrent: not implemented")
	return model.Track{}, false
}

// ─── Navigation ──────────────────────────────────────────────────────────────

// Next advances Curr to the next node and returns it.
// Returns (nil, false) when already at the tail.
//
// TODO: implement O(1) — just follow the Next pointer
func (p *Playlist) Next() (*model.Node, bool) {
	// panic("Next: not implemented")
	return nil, false
}

// Prev moves Curr to the previous node and returns it.
// Returns (nil, false) when already at the head.
//
// TODO: implement O(1) — just follow the Prev pointer
func (p *Playlist) Prev() (*model.Node, bool) {
	// panic("Prev: not implemented")
	return nil, false
}

// JumpTo sets Curr to the node whose Track.ID equals id.
// Returns (nil, false) if no such node exists.
//
// TODO: implement — linear scan from Head; O(n) is acceptable here
func (p *Playlist) JumpTo(id int) (*model.Node, bool) {
	// panic("JumpTo: not implemented")
	return nil, false
}

// ─── Query helpers ────────────────────────────────────────────────────────────

// Current returns the current node without moving Curr.
func (p *Playlist) Current() *model.Node {
	return p.Curr
}

// IsEmpty reports whether the playlist has no tracks.
func (p *Playlist) IsEmpty() bool {
	return p.Size == 0
}

// ToSlice returns all tracks in order (head → tail).
// Used by the UI to render the list — do NOT store the result
// and treat it as the source of truth; the linked list is.
//
// TODO: implement — walk from Head to Tail, append each Track
func (p *Playlist) ToSlice() []model.Track {
	// panic("ToSlice: not implemented")
	return []model.Track{}
}

// AtHead reports whether Curr is the first node.
func (p *Playlist) AtHead() bool {
	return p.Curr != nil && p.Curr == p.Head
}

// AtTail reports whether Curr is the last node.
func (p *Playlist) AtTail() bool {
	return p.Curr != nil && p.Curr == p.Tail
}

func (p *Playlist) Traverse() {
	cur := p.Head
	for cur != nil {
		fmt.Println(cur.Track.Title)
		cur = cur.Next
	}
}
