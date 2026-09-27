package scanner

import (
	"context"
)

// Walk traverses the whole tree and sorts entries into MP3 files, skipped files and errors.
//
// It returns a non-nil error only when the walk as a whole must stop (ctx cancelled).
// Problems with individual entries go into WalkResult.Errors instead.
func (s *Scanner) Walk(ctx context.Context) (WalkResult, error) {
	// TODO(you): implement with fs.WalkDir(s.fsys, ".", fn). Covered by TestWalk*.
	//
	// Rules:
	//   1. Regular file with a .mp3 extension (any case) → Files; Size/ModTime come from d.Info().
	//   2. Any other regular file → Skipped.
	//   3. Hidden directory (name starts with ".") → don't descend: return fs.SkipDir.
	//      Careful: the root itself is named "." and is NOT hidden.
	//   4. Callback gets err != nil (unreadable entry) → append FileError{path, err}, keep walking.
	//   5. ctx cancelled → abort and return an error that wraps ctx.Err().
	//
	// Hints:
	//   - The callback type is fs.WalkDirFunc: func(path string, d fs.DirEntry, err error) error.
	//     Returning a non-nil error aborts the walk. fs.SkipDir is a special sentinel
	//     meaning "don't descend into this directory".
	//   - For an unreadable directory WalkDir calls the callback TWICE: first with err == nil
	//     (the entry itself), then again with the ReadDir error. When err != nil, d may be nil
	//     (for the root), so check err before touching d.
	//   - WalkDir visits entries in lexical order, so Files come out sorted for free.
	//   - Checking ctx: `if err := ctx.Err(); err != nil { return err }` at the top of the callback.
	//     WalkDir returns the callback's error unchanged, so wrap it once afterwards:
	//     fmt.Errorf("walk: %w", err).
	//   - LEARN: a closure can append to `res` declared outside of it. No ref/out params needed.
	return WalkResult{}, errTODO
}

// isMP3 reports whether name has a .mp3 extension, case-insensitively.
func isMP3(name string) bool {
	// TODO(you): covered by TestIsMP3. See path.Ext and strings.EqualFold.
	//
	// LEARN: use package `path` (always '/') for fs.FS paths and `path/filepath`
	// (OS separator) for real OS paths. fs.FS paths are always slash-separated, even on Windows.
	_ = name
	return false
}
