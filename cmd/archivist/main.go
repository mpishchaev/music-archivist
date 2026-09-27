// scaffold: entry point. Keep main tiny — all logic lives in internal/.
//
// LEARN: `package main` + `func main()` is the only executable entry point.
// There is no Program class and no `static void Main(string[] args)`: args come from os.Args,
// and the exit code is set explicitly with os.Exit (like returning int from Main).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mpishchaev/music-archivist/internal/cli"
)

func main() {
	// LEARN: context.Context ~ CancellationToken, but it also carries deadlines and
	// request-scoped values. NotifyContext cancels ctx on Ctrl+C / SIGTERM
	// (roughly Console.CancelKeyPress + a CancellationTokenSource).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// LEARN: defer ~ finally: runs when the surrounding function returns, LIFO order.
	defer stop()

	if err := cli.NewRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		// LEARN: os.Exit skips deferred calls, so call stop() by hand before exiting.
		stop()
		os.Exit(1)
	}
}
