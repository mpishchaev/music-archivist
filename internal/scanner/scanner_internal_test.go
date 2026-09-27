// LEARN: `package scanner` (same package) = white-box test: sees unexported identifiers
// like isMP3 and hashFile. Like [InternalsVisibleTo], but per file and with zero config.
// Compare with scanner_test.go, which uses `package scanner_test`.
package scanner

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMP3(t *testing.T) {
	t.Parallel()

	// LEARN: table-driven test ≈ xUnit [Theory] + [InlineData], but the table is plain data:
	// a slice of anonymous structs. Adding a case = adding one line.
	tests := []struct {
		name string
		want bool
	}{
		{"song.mp3", true},
		{"SONG.MP3", true},
		{"Song.Mp3", true},
		{"dir/sub/song.mp3", true},
		{".mp3", true},
		{"song.mp4", false},
		{"song.mp3.part", false},
		{"song.mp3 ", false},
		{"mp3", false},
		{"", false},
	}

	for _, tc := range tests {
		// LEARN: t.Run creates a named subtest: `go test -run 'TestIsMP3/SONG'` runs one case.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isMP3(tc.name))
		})
	}
}

func TestHashFile(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"empty.mp3": {Data: []byte{}},
		"hello.mp3": {Data: []byte("hello")},
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr error // nil = expect success
	}{
		{
			name: "empty file",
			path: "empty.mp3",
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "known content",
			path: "hello.mp3",
			want: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:    "missing file",
			path:    "nope.mp3",
			wantErr: fs.ErrNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := hashFile(fsys, tc.path)

			if tc.wantErr != nil {
				// LEARN: ErrorIs walks the %w chain, so your "hash %s: %w" wrapping must
				// preserve the original fs.ErrNotExist.
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
