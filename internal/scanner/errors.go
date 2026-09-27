package scanner

// FileError is a failure tied to one file or directory. Scanning continues past it;
// all FileErrors end up in the final report.
//
// LEARN: any type with an `Error() string` method satisfies the built-in `error` interface.
// Adding `Unwrap() error` plugs it into errors.Is / errors.As chains, much like
// InnerException, except matching is by value/type instead of by catch clauses.
//
// LEARN: value receivers here (not *FileError): the struct is small and immutable, and we
// store it by value in slices. With value receivers both FileError and *FileError
// satisfy `error`.
type FileError struct {
	Path string
	Err  error
}

// Error formats as "<path>: <cause>", e.g. "music/a.mp3: permission denied".
func (e FileError) Error() string {
	// TODO(you): covered by TestFileError. Hint: fmt.Sprintf or string concatenation.
	return ""
}

// Unwrap returns the underlying cause so errors.Is(fe, fs.ErrPermission) works.
func (e FileError) Unwrap() error {
	// TODO(you): covered by TestFileError.
	return nil
}
