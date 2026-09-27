// scaffold: cobra wiring. Commands stay thin: parse flags → build services → call them.
//
// LEARN: `internal/` is enforced by the compiler: packages under it can be imported only
// from within this module (like `internal` visibility at assembly level in C#).
package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// errNotImplemented is returned by commands whose milestone isn't done yet.
//
// LEARN: a "sentinel error" — a package-level value compared with errors.Is.
// Lower-case first letter = unexported (private to the package). There are no
// public/private keywords in Go: capitalization IS the access modifier.
var errNotImplemented = errors.New("not implemented yet")

// globalOptions holds flags shared by all subcommands.
//
// LEARN: plain struct instead of IOptions<T>/DI. It's passed explicitly by pointer
// so subcommands see values after cobra parses flags.
type globalOptions struct {
	configPath string
	dbPath     string
	logLevel   string
	workers    int

	logger *slog.Logger // built in PersistentPreRunE from logLevel
}

// NewRootCmd builds the root command with all subcommands attached.
//
// LEARN: a constructor is just a function named NewX. No `new` keyword ceremony,
// no DI container — dependencies are passed as arguments.
func NewRootCmd() *cobra.Command {
	opts := &globalOptions{}

	root := &cobra.Command{
		Use:           "archivist",
		Short:         "Scan, deduplicate and organize an MP3 collection",
		SilenceUsage:  true, // don't dump usage on runtime errors
		SilenceErrors: true, // main prints the error itself
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			logger, err := newLogger(opts.logLevel)
			if err != nil {
				return err
			}
			opts.logger = logger
			return nil
		},
	}

	f := root.PersistentFlags()
	f.StringVar(&opts.configPath, "config", "archivist.yaml", "path to YAML config")
	f.StringVar(&opts.dbPath, "db", "archivist.db", "path to SQLite index")
	f.StringVar(&opts.logLevel, "log-level", "info", "debug|info|warn|error")
	f.IntVar(&opts.workers, "workers", 4, "number of parallel workers")

	root.AddCommand(
		newScanCmd(opts),
		newDupesCmd(opts),
		newPlanCmd(opts),
		newApplyCmd(opts),
	)
	return root
}

func newLogger(level string) (*slog.Logger, error) {
	var lvl slog.Level
	// LEARN: UnmarshalText is part of the encoding.TextUnmarshaler interface —
	// slog.Level satisfies it implicitly, nobody wrote `: ITextUnmarshaler`.
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		// LEARN: %w wraps the error so callers can still errors.Is/As the cause
		// (like InnerException, but checked by value/type, not by catch blocks).
		return nil, fmt.Errorf("invalid --log-level %q: %w", level, err)
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(h), nil
}
