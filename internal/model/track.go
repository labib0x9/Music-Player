package model

import "time"

// Track represents a single audio file in the library.
type Track struct {
	ID       int
	Title    string
	Artist   string
	Album    string
	Path     string
	Duration time.Duration
}

// Node is a doubly-linked list node wrapping a Track.
// ─────────────────────────────────────────────────────
// TODO (your task): implement the linked-list logic in
//
//	internal/playlist/playlist.go
//
// The fields below are the contract every other package
// relies on — do NOT change them.
type Node struct {
	Track Track
	Prev  *Node
	Next  *Node
}
