// Package scanner discovers MP3 files in a source tree and hashes their content.
//
// Usage: New(fsys, workers) → Walk → Hash.
// The scanner only ever reads from fsys; it never writes.
package scanner

import (
	"errors"
	"io/fs"
	"time"
)

// errTODO marks functions you still have to implement. Delete it when the milestone is done.
var errTODO = errors.New("TODO(you): not implemented")

// File is an MP3 candidate discovered by Walk, not hashed yet.
type File struct {
	Path    string // slash-separated, relative to the scanned fs.FS root
	Size    int64
	ModTime time.Time
}

// WalkResult is everything Walk found.
type WalkResult struct {
	Files   []File      // MP3 files, in lexical order
	Skipped []string    // non-MP3 regular files: reported, never touched
	Errors  []FileError // entries that could not be read; the walk continued past them
}

// Scanner walks and hashes one source tree.
//
// LEARN: fields are unexported, so the only way to build a valid Scanner is New.
// That's Go's substitute for a constructor with invariants.
type Scanner struct {
	fsys    fs.FS
	workers int
}

// New returns a Scanner over fsys that hashes with up to workers goroutines (min 1).
//
// LEARN: fs.FS is a one-method interface (Open). os.DirFS("/home/me/Music") gives you
// the real disk, fstest.MapFS gives you an in-memory tree for tests, with no mocking library.
// This is "accept interfaces, return structs": New takes an interface, returns *Scanner.
func New(fsys fs.FS, workers int) *Scanner {
	if workers < 1 {
		workers = 1
	}
	return &Scanner{fsys: fsys, workers: workers}
}
