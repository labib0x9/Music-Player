// Package utils provides small helpers used across the project.
package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yourname/music-player/internal/model"
)

var supportedExts = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".ogg":  true,
}

// ScanDir walks dir recursively and returns a slice of Tracks for every
// supported audio file found.  IDs are assigned sequentially starting at 1.
func ScanDir(dir string) ([]model.Track, error) {
	var tracks []model.Track
	id := 1

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}
		title := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		tracks = append(tracks, model.Track{
			ID:     id,
			Title:  title,
			Artist: "Unknown",
			Path:   path,
		})
		id++
		return nil
	})
	return tracks, err
}
