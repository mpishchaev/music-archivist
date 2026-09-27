// Test doubles for fs.FS. Hand-written fakes instead of a mocking framework.
//
// LEARN: *_test.go files are compiled only by `go test`, so these helpers never ship.
package scanner_test

import (
	"io/fs"
	"sync/atomic"
	"time"
)

// faultyFS wraps an fs.FS and fails Open for selected paths.
//
// LEARN: embedding an interface (`fs.FS` without a field name) "inherits" its methods;
// we override Open only. Side effect worth knowing: the wrapper hides MapFS's extra
// methods (ReadDir, Stat), so fs.WalkDir falls back to Open for directories too,
// which is exactly what lets us simulate an unreadable directory.
type faultyFS struct {
	fs.FS
	failOn map[string]error
}

func (f faultyFS) Open(name string) (fs.File, error) {
	if err, ok := f.failOn[name]; ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	return f.FS.Open(name)
}

// countingFS tracks how many files are open at the same time and remembers the peak.
// If a file is never closed, `cur` never goes down and the peak exceeds the worker
// limit, so this also catches leaked file handles.
type countingFS struct {
	fs.FS
	cur  atomic.Int32
	peak atomic.Int32
}

func (c *countingFS) Open(name string) (fs.File, error) {
	f, err := c.FS.Open(name)
	if err != nil {
		return nil, err
	}

	n := c.cur.Add(1)
	// LEARN: lock-free "max" via compare-and-swap (Interlocked.CompareExchange loop).
	for {
		p := c.peak.Load()
		if n <= p || c.peak.CompareAndSwap(p, n) {
			break
		}
	}
	time.Sleep(5 * time.Millisecond) // widen the window so parallel opens actually overlap

	return &countingFile{File: f, cur: &c.cur}, nil
}

type countingFile struct {
	fs.File
	cur *atomic.Int32
}

func (f *countingFile) Close() error {
	f.cur.Add(-1)
	return f.File.Close()
}
