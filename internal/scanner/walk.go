package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Walk traverses the whole tree and sorts entries into MP3 files, skipped files and errors.
//
// It returns a non-nil error only when the walk as a whole must stop (ctx cancelled).
// Problems with individual entries go into WalkResult.Errors instead.
func (s *Scanner) Walk(ctx context.Context) (WalkResult, error) {
	var res WalkResult

	// LEARN: the callback is a closure: it appends to `res` from the enclosing function,
	// no ref/out params needed. Its return value steers the walk: nil = continue,
	// fs.SkipDir = don't descend, any other error = abort (WalkDir returns it unchanged).
	err := fs.WalkDir(s.fsys, ".", func(p string, d fs.DirEntry, err error) error {
		// 1. Cancellation first: a non-nil return aborts the whole walk.
		if err := ctx.Err(); err != nil {
			return err
		}

		// 2. Unreadable entry: record and keep walking. Don't touch d here, it may be nil.
		// For an unreadable directory WalkDir calls us twice: first with err == nil, then
		// with the ReadDir error, so each bad directory yields exactly one FileError.
		if err != nil {
			res.Errors = append(res.Errors, FileError{Path: p, Err: err})
			return nil
		}

		// 3. Directories: skip hidden ones, but never the root ".".
		if d.IsDir() {
			if p != "." && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}

		// 4. Only regular files from here on (ignore symlinks, devices...).
		if !d.Type().IsRegular() {
			return nil
		}

		// 5. Non-MP3 → Skipped.
		if !isMP3(p) {
			res.Skipped = append(res.Skipped, p)
			return nil
		}

		// 6. MP3 → need size and mtime from Info(), which can fail.
		info, err := d.Info()
		if err != nil {
			res.Errors = append(res.Errors, FileError{Path: p, Err: err})
			return nil
		}

		res.Files = append(res.Files, File{Path: p, Size: info.Size(), ModTime: info.ModTime()})
		return nil
	})
	if err != nil {
		return WalkResult{}, fmt.Errorf("walk: %w", err)
	}
	return res, nil
}

// isMP3 reports whether name has a .mp3 extension, case-insensitively.
//
// LEARN: use package `path` (always '/') for fs.FS paths and `path/filepath`
// (OS separator) for real OS paths. fs.FS paths are always slash-separated, even on Windows.
func isMP3(name string) bool {
	return strings.EqualFold(path.Ext(name), ".mp3")
}
