// Package model holds plain data types shared between packages.
//
// LEARN: no behavior here, just data (like DTOs / records). Keeping it dependency-free
// lets every other package import it without creating import cycles — Go forbids
// cyclic imports at compile time, unlike .NET which only forbids cyclic project refs.
package model

import "time"

// Track is one MP3 file found under a source root.
// Fields are added milestone by milestone (tags in M2, duration/bitrate in M7).
type Track struct {
	Root    string    // absolute OS path of the source root the file was found under
	RelPath string    // slash-separated path relative to Root (fs.FS convention)
	Size    int64     // bytes
	ModTime time.Time // last modification time, used for incremental rescans (M3)
	SHA256  string    // hex-encoded content hash, used for exact dedup (M5)
}
