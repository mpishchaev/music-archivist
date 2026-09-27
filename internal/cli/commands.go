// scaffold: subcommand stubs. Each milestone replaces a stub body with real wiring.
package cli

import (
	"github.com/spf13/cobra"
)

func newScanCmd(opts *globalOptions) *cobra.Command {
	// LEARN: variables captured by the closure below (srcDirs, opts) live as long as the
	// closure does — same semantics as C# lambdas capturing locals.
	var srcDirs []string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Walk source dirs, hash MP3s, read tags and store them in the index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.logger.Info("scan", "src", srcDirs, "workers", opts.workers)
			return errNotImplemented // milestone 1–3
		},
	}
	cmd.Flags().StringSliceVar(&srcDirs, "src", nil, "source directory (repeatable)")
	_ = cmd.MarkFlagRequired("src")
	return cmd
}

func newDupesCmd(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "dupes",
		Short: "Print duplicate groups found in the index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 5
		},
	}
}

func newPlanCmd(opts *globalOptions) *cobra.Command {
	var dstDir string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Compute target paths for every kept track",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 4
		},
	}
	cmd.Flags().StringVar(&dstDir, "dst", "", "target library root")
	_ = cmd.MarkFlagRequired("dst")
	return cmd
}

func newApplyCmd(opts *globalOptions) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Copy files according to the plan",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errNotImplemented // milestone 4
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print actions without touching the disk")
	return cmd
}
