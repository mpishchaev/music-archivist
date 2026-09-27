// LEARN: `package scanner_test` = black-box test: only the exported API is visible,
// exactly like a consumer sees it. Prefer this for public behavior; use the internal
// test file for unexported helpers.
package scanner_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mpishchaev/music-archivist/internal/model"
	"github.com/mpishchaev/music-archivist/internal/scanner"
)

// ---------- FileError ----------

func TestFileError(t *testing.T) {
	t.Parallel()

	var err error = scanner.FileError{Path: "music/a.mp3", Err: fs.ErrPermission}

	assert.EqualError(t, err, "music/a.mp3: permission denied")
	assert.ErrorIs(t, err, fs.ErrPermission, "Unwrap must expose the cause")

	// errors.As digs a FileError out of an arbitrarily wrapped chain.
	wrapped := fmt.Errorf("scan: %w", err)
	var fe scanner.FileError
	require.ErrorAs(t, wrapped, &fe)
	assert.Equal(t, "music/a.mp3", fe.Path)
}

// ---------- Walk ----------

func TestWalk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fsys        fstest.MapFS
		wantFiles   []string
		wantSkipped []string
	}{
		{
			name:        "empty tree",
			fsys:        fstest.MapFS{},
			wantFiles:   nil,
			wantSkipped: nil,
		},
		{
			name: "mp3 in any case, others skipped",
			fsys: fstest.MapFS{
				"a.mp3":     {},
				"B.MP3":     {},
				"cover.jpg": {},
				"notes.txt": {},
			},
			// Lexical byte order: uppercase letters sort before lowercase.
			wantFiles:   []string{"B.MP3", "a.mp3"},
			wantSkipped: []string{"cover.jpg", "notes.txt"},
		},
		{
			name: "nested directories",
			fsys: fstest.MapFS{
				"Artist/Album/01.mp3": {},
				"Artist/Album/02.mp3": {},
				"Artist/single.mp3":   {},
				"loose.flac":          {},
			},
			wantFiles:   []string{"Artist/Album/01.mp3", "Artist/Album/02.mp3", "Artist/single.mp3"},
			wantSkipped: []string{"loose.flac"},
		},
		{
			name: "hidden directories are not entered",
			fsys: fstest.MapFS{
				".git/objects/x.mp3":   {},
				"music/.cache/y.mp3":   {},
				"music/visible.mp3":    {},
				".Trash-1000/junk.txt": {},
			},
			wantFiles:   []string{"music/visible.mp3"},
			wantSkipped: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := scanner.New(tc.fsys, 1).Walk(t.Context())

			require.NoError(t, err)
			assertSameStrings(t, tc.wantFiles, filePaths(res.Files))
			assertSameStrings(t, tc.wantSkipped, res.Skipped)
			assert.Empty(t, res.Errors)
		})
	}
}

func TestWalk_FillsSizeAndModTime(t *testing.T) {
	t.Parallel()

	mtime := time.Date(2019, 5, 17, 10, 30, 0, 0, time.UTC)
	fsys := fstest.MapFS{
		"song.mp3": {Data: []byte("12345"), ModTime: mtime},
	}

	res, err := scanner.New(fsys, 1).Walk(t.Context())

	require.NoError(t, err)
	require.Len(t, res.Files, 1)
	assert.Equal(t, scanner.File{Path: "song.mp3", Size: 5, ModTime: mtime}, res.Files[0])
}

func TestWalk_UnreadableDirectoryIsReportedNotFatal(t *testing.T) {
	t.Parallel()

	fsys := faultyFS{
		FS: fstest.MapFS{
			"ok/a.mp3":     {},
			"broken/b.mp3": {},
			"z.mp3":        {},
		},
		failOn: map[string]error{"broken": fs.ErrPermission},
	}

	res, err := scanner.New(fsys, 1).Walk(t.Context())

	require.NoError(t, err, "one bad directory must not abort the walk")
	assert.Equal(t, []string{"ok/a.mp3", "z.mp3"}, filePaths(res.Files))
	require.Len(t, res.Errors, 1)
	assert.Equal(t, "broken", res.Errors[0].Path)
	assert.ErrorIs(t, res.Errors[0], fs.ErrPermission)
}

func TestWalk_Cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // already cancelled before we start

	_, err := scanner.New(fstest.MapFS{"a.mp3": {}}, 1).Walk(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

// ---------- Hash ----------

func TestHash_OrderAndContent(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}
	var files []scanner.File
	var want []model.Track
	mtime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range 12 { // LEARN: Go 1.22+ range over int: i = 0..11
		name := fmt.Sprintf("track%02d.mp3", i)
		data := []byte(fmt.Sprintf("content of %d", i))
		fsys[name] = &fstest.MapFile{Data: data, ModTime: mtime}

		files = append(files, scanner.File{Path: name, Size: int64(len(data)), ModTime: mtime})
		want = append(want, model.Track{
			RelPath: name,
			Size:    int64(len(data)),
			ModTime: mtime,
			SHA256:  sha256Hex(data),
		})
	}

	tracks, failures, err := scanner.New(fsys, 4).Hash(t.Context(), files)

	require.NoError(t, err)
	assert.Empty(t, failures)
	assert.Equal(t, want, tracks, "same order as input, Root left empty")
}

func TestHash_PartialFailure(t *testing.T) {
	t.Parallel()

	fsys := faultyFS{
		FS: fstest.MapFS{
			"a.mp3": {Data: []byte("a")},
			"b.mp3": {Data: []byte("b")},
			"c.mp3": {Data: []byte("c")},
		},
		failOn: map[string]error{"b.mp3": fs.ErrPermission},
	}
	files := []scanner.File{{Path: "a.mp3"}, {Path: "b.mp3"}, {Path: "c.mp3"}}

	tracks, failures, err := scanner.New(fsys, 2).Hash(t.Context(), files)

	require.NoError(t, err, "a single bad file is not a fatal error")
	assert.Equal(t, []string{"a.mp3", "c.mp3"}, trackPaths(tracks))
	require.Len(t, failures, 1)
	assert.Equal(t, "b.mp3", failures[0].Path)
	assert.ErrorIs(t, failures[0], fs.ErrPermission)
}

func TestHash_RespectsWorkerLimit(t *testing.T) {
	t.Parallel()

	const workers = 3
	mfs := fstest.MapFS{}
	var files []scanner.File
	for i := range 20 {
		name := fmt.Sprintf("%02d.mp3", i)
		mfs[name] = &fstest.MapFile{Data: []byte(name)}
		files = append(files, scanner.File{Path: name})
	}
	cfs := &countingFS{FS: mfs}

	_, _, err := scanner.New(cfs, workers).Hash(t.Context(), files)

	require.NoError(t, err)
	peak := cfs.peak.Load()
	assert.LessOrEqual(t, peak, int32(workers), "more files open at once than workers (or a file was never closed)")
	assert.GreaterOrEqual(t, peak, int32(2), "files were hashed sequentially; expected parallelism")
	assert.Zero(t, cfs.cur.Load(), "some files were not closed")
}

func TestHash_Cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	fsys := fstest.MapFS{"a.mp3": {Data: []byte("a")}}
	_, _, err := scanner.New(fsys, 2).Hash(ctx, []scanner.File{{Path: "a.mp3"}})

	require.ErrorIs(t, err, context.Canceled)
}

// ---------- benchmark ----------

// Run: go test -run '^$' -bench Hash -benchmem ./internal/scanner
func BenchmarkHash(b *testing.B) {
	fsys := fstest.MapFS{}
	var files []scanner.File
	payload := make([]byte, 1<<20) // 1 MiB each
	for i := range 32 {
		name := fmt.Sprintf("%02d.mp3", i)
		fsys[name] = &fstest.MapFile{Data: payload}
		files = append(files, scanner.File{Path: name})
	}

	for _, workers := range []int{1, 2, 4, 8} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			s := scanner.New(fsys, workers)
			b.SetBytes(int64(len(payload) * len(files)))
			// LEARN: b.Loop (Go 1.24+) replaces `for i := 0; i < b.N; i++` and keeps setup out of timing.
			for b.Loop() {
				if _, _, err := s.Hash(b.Context(), files); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------- helpers ----------

// assertSameStrings treats nil and an empty slice as equal.
//
// LEARN: a nil slice and []string{} both have len 0 and both work with range/append,
// but reflect.DeepEqual (and so assert.Equal) tells them apart. Idiomatic Go code
// usually returns nil for "nothing" (`var out []string`), but callers shouldn't care.
func assertSameStrings(t *testing.T, want, got []string) {
	t.Helper() // LEARN: failures are reported at the caller's line, not here
	if len(want) == 0 {
		assert.Empty(t, got)
		return
	}
	assert.Equal(t, want, got)
}

// LEARN: no LINQ `.Select(f => f.Path)`, just a loop. `var out []string` starts nil,
// which lets assert.Equal(nil, ...) match an empty result.
func filePaths(files []scanner.File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.Path)
	}
	return out
}

func trackPaths(tracks []model.Track) []string {
	var out []string
	for _, t := range tracks {
		out = append(out, t.RelPath)
	}
	return out
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
