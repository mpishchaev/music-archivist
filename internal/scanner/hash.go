package scanner

import (
	"context"
	"io/fs"

	"github.com/mpishchaev/music-archivist/internal/model"
)

// Hash computes SHA-256 for every file using up to s.workers goroutines in parallel.
//
// Contract (covered by TestHash*):
//   - tracks keep the order of `files`; failed files are simply absent from tracks;
//   - a file that fails to open or read → one FileError in failures, the rest is still hashed;
//   - never more than s.workers files are open at the same time;
//   - ctx cancelled → return an error wrapping ctx.Err() (partial results may be dropped);
//   - Track.Root stays empty: the caller knows which root this Scanner covers.
func (s *Scanner) Hash(ctx context.Context, files []File) (tracks []model.Track, failures []FileError, err error) {
	// TODO(you): implement. Read docs/notes/01-concurrency.md first.
	//
	// Pick one shape (both are idiomatic, try both if you like):
	//   A) errgroup: g, ctx := errgroup.WithContext(ctx); g.SetLimit(s.workers);
	//      for i := range files { g.Go(func() error { ... }) }; err := g.Wait()
	//   B) classic worker pool: a jobs channel of indexes, s.workers goroutines doing
	//      `for i := range jobs`, a sync.WaitGroup, and close(jobs) once everything is sent.
	//
	// Tips:
	//   - Preallocate `results := make([]result, len(files))`. Goroutines writing to DIFFERENT
	//     indexes of a slice do not race, so no mutex is needed. Compact into tracks/failures
	//     afterwards; that also gives you the ordering guarantee.
	//   - A per-file failure must NOT be returned from g.Go: with errgroup.WithContext that
	//     would cancel everyone else. Store it in results[i] and return nil.
	//   - Cancellation: nothing stops by itself. Check ctx.Err() before starting each file.
	//   - GOTCHA (Claude fell into it while validating these tests): the ctx returned by
	//     errgroup.WithContext is cancelled as soon as g.Wait() returns, even on success.
	//     Don't shadow: write `g, gctx := errgroup.WithContext(ctx)`, use gctx inside
	//     the goroutines, and check the ORIGINAL ctx after Wait.
	//   - Go 1.22+: the loop variable `i` is per-iteration, so capturing it in a closure is safe
	//     (unlike the old C# 4 foreach-closure gotcha, and unlike Go before 1.22).
	//
	// The named results (tracks, failures, err) are just documentation here. Plain `return x, y, z` is fine.
	return nil, nil, errTODO
}

// hashFile streams one file through SHA-256 and returns the lowercase hex digest.
func hashFile(fsys fs.FS, path string) (string, error) {
	// TODO(you): covered by TestHashFile.
	//   fsys.Open → defer f.Close() → io.Copy(h, f) with h := sha256.New() → hex.EncodeToString(h.Sum(nil))
	//
	// LEARN: io.Copy streams in 32 KB chunks: memory stays flat whatever the file size
	// (compare Stream.CopyTo vs File.ReadAllBytes). hash.Hash is an io.Writer, which is why
	// io.Copy can target it: interfaces compose through tiny shared contracts.
	// Wrap errors with context: fmt.Errorf("hash %s: %w", path, err).
	//
	// The linter (errcheck) will flag a bare `defer f.Close()`. For a file opened READ-ONLY,
	// a Close error can't lose data, so ignoring it is accepted practice. Make it explicit:
	// `defer func() { _ = f.Close() }()`. For files you WRITE (M4) it's the opposite:
	// Close can report a failed flush and must be checked.
	_, _ = fsys, path
	return "", errTODO
}
