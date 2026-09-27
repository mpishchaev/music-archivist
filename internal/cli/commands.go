package cli

// scaffold: subcommands. Each milestone replaces a stub body with real wiring.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mpishchaev/music-archivist/internal/scanner"
)

func newScanCmd(opts *globalOptions) *cobra.Command {
	// LEARN: variables captured by the closure below (srcDirs, opts) live as long as the
	// closure does — same semantics as C# lambdas capturing locals.
	var srcDirs []string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Walk source dirs, hash MP3s, read tags and store them in the index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// TODO(M3): store tracks in the SQLite index instead of just printing a summary.
			out := cmd.OutOrStdout()
			for _, src := range srcDirs {
				root, err := filepath.Abs(src)
				if err != nil {
					return fmt.Errorf("resolve %s: %w", src, err)
				}
				opts.logger.Info("scanning", "root", root, "workers", opts.workers)

				// LEARN: os.DirFS is read-only by construction: fs.FS has no Write/Create.
				// The type system itself guarantees we never modify the sources.
				s := scanner.New(os.DirFS(root), opts.workers)

				walked, err := s.Walk(cmd.Context())
				if err != nil {
					return err
				}
				tracks, failures, err := s.Hash(cmd.Context(), walked.Files)
				if err != nil {
					return err
				}
				for i := range tracks {
					tracks[i].Root = root
				}

				for _, fe := range append(walked.Errors, failures...) {
					opts.logger.Warn("unreadable", "path", fe.Path, "err", fe.Err)
				}
				fmt.Fprintf(out, "%s: %d mp3 hashed, %d skipped (non-mp3), %d errors\n",
					root, len(tracks), len(walked.Skipped), len(walked.Errors)+len(failures))
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&srcDirs, "src", nil, "source directory (repeatable)")
	_ = cmd.MarkFlagRequired("src")
	return cmd
}

func newDupesCmd(_ *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "dupes",
		Short: "Print duplicate groups found in the index",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 5
		},
	}
}

func newPlanCmd(_ *globalOptions) *cobra.Command {
	var dstDir string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Compute target paths for every kept track",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 4
		},
	}
	cmd.Flags().StringVar(&dstDir, "dst", "", "target library root")
	_ = cmd.MarkFlagRequired("dst")
	return cmd
}

func newApplyCmd(_ *globalOptions) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Copy files according to the plan",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 4
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print actions without touching the disk")
	return cmd
}
