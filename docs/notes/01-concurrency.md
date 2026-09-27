# 01 — Concurrency, context, fs.FS, streaming I/O

Everything milestone 1 (scanner) uses. Examples are deliberately **not** the scanner code:
they show the pattern on toy problems, so you still get to write the real thing.

## Mental model: no async/await

| .NET | Go |
|---|---|
| `async Task<T> Foo()` + `await` | Plain blocking `func Foo() (T, error)`. The runtime parks the goroutine during I/O and runs others on the same OS thread. No "function coloring". |
| `Task.Run(() => Work())` | `go work()`: fire-and-forget. No handle, no result, no exception propagation. |
| Thread pool | The Go scheduler multiplexes goroutines (a few KB of stack each) onto `GOMAXPROCS` OS threads. 100k goroutines are fine. |

Because `go f()` returns nothing, **you** decide how results and errors come back:
channels, a shared slice with distinct indexes, or `errgroup`.

## Goroutines + WaitGroup (≈ Task.WhenAll)

```go
var wg sync.WaitGroup
for _, url := range urls {
    wg.Go(func() {           // Go 1.25+: wg.Go = wg.Add(1) + go + defer wg.Done()
        fetch(url)            // Go 1.22+: `url` is a fresh variable each iteration, safe to capture
    })
}
wg.Wait()
```

## Channels (≈ System.Threading.Channels)

```go
ch := make(chan int)      // unbuffered: send blocks until someone receives (a rendezvous)
ch := make(chan int, 10)  // buffered: send blocks only when 10 items are waiting

ch <- 42                  // send
v := <-ch                 // receive
v, ok := <-ch             // ok == false → channel closed and drained
close(ch)                 // only the SENDER closes; it means "no more values"
for v := range ch { }     // loops until closed (≈ await foreach over ReadAllAsync)
```

Rules that bite:
- Sending to a closed channel → **panic**. Closing twice → panic.
- A receive from a nil channel blocks forever. Good for disabling a `select` case, bad by accident.
- A goroutine blocked forever on a channel is a **leak** (never collected). Always have an exit path.

## select (≈ Task.WhenAny over channels)

```go
select {
case v := <-results:
    use(v)
case <-ctx.Done():        // cancellation
    return ctx.Err()
case <-time.After(time.Second):
    return errTimeout
}
```

## context.Context (≈ CancellationToken + more)

- Always the **first parameter**, named `ctx`. Never stored in a struct.
- `ctx.Done()` is a channel closed on cancel. `ctx.Err()` returns `context.Canceled` / `context.DeadlineExceeded`, or nil.
- Derive children: `context.WithCancel`, `WithTimeout`, `WithDeadline`. Always `defer cancel()`.
- Cancellation is **cooperative**, as in .NET: nothing stops by itself, and code must check `ctx.Err()` or select on `ctx.Done()`.
- In tests: `t.Context()` (Go 1.24+) is cancelled automatically when the test ends.

## Worker pool, two idiomatic shapes

### A. Channel of jobs + N workers

```go
jobs := make(chan string)
out := make(chan int)

var wg sync.WaitGroup
for range workers {                       // start N workers
    wg.Go(func() {
        for word := range jobs {          // each worker pulls until jobs is closed
            out <- len(word)
        }
    })
}

go func() {                               // producer
    defer close(jobs)
    for _, w := range words {
        select {
        case jobs <- w:
        case <-ctx.Done():
            return
        }
    }
}()

go func() { wg.Wait(); close(out) }()     // close out after the last worker finishes

for n := range out { total += n }         // consumer
```
Good when results stream out and order doesn't matter.

### B. errgroup with a limit (≈ Parallel.ForEachAsync with MaxDegreeOfParallelism)

```go
g, gctx := errgroup.WithContext(ctx)
g.SetLimit(workers)                  // g.Go blocks while `workers` goroutines are running

lengths := make([]int, len(words))   // one slot per input: distinct indexes → no data race
for i, w := range words {
    g.Go(func() error {
        if err := gctx.Err(); err != nil {
            return err
        }
        lengths[i] = len(w)
        return nil                   // returning non-nil cancels gctx for everyone
    })
}
if err := g.Wait(); err != nil {     // first non-nil error from any g.Go
    return err
}
```
Good when you have a known slice of inputs and want results in input order.

> **Gotcha:** `gctx` is cancelled when `Wait` returns, **even on success**. If you write
> `g, ctx := errgroup.WithContext(ctx)` (shadowing) and check `ctx.Err()` after `Wait`,
> you'll always see `context canceled`. Keep two names.

> **Design choice:** errgroup's "first error cancels all" suits *fatal* errors.
> A per-item failure (one unreadable file) is **data**, not a fatal error: record it and return nil.

## Data races and `-race`

- Two goroutines touching the same memory, with at least one writing and no synchronization, is a data race and undefined behavior.
- Writing to *different* elements of one slice is **not** a race. `append` to a shared slice from several goroutines **is**.
- A map written concurrently is a race, and the runtime may even crash with `concurrent map writes`.
- Tools: `sync.Mutex` (≈ `lock`), `sync/atomic` (≈ `Interlocked`), or better, don't share and communicate via channels.
- Always test with `go test -race`. It instruments memory accesses and reports real races with stack traces.

## fs.FS: an abstraction you get for free

```go
type FS interface { Open(name string) (fs.File, error) }
```
- `os.DirFS("/music")` is the real disk, **read-only by type**: there's no Write in the interface.
- `fstest.MapFS{"a/b.mp3": {Data: []byte("x")}}` is an in-memory tree for tests.
- `fs.WalkDir(fsys, ".", fn)` walks recursively in lexical order. `fs.ReadFile`, `fs.Stat` and `fs.Glob` also exist.
- Paths inside an fs.FS are always slash-separated, unrooted, and use no `..`. Use package `path`, not `filepath`.
- Compare with .NET's `System.IO.Abstractions`: same idea, but it's in the standard library and one method wide.

## Streaming I/O

```go
f, err := fsys.Open(name)
if err != nil { return err }
defer func() { _ = f.Close() }()

n, err := io.Copy(dst, f)   // dst is any io.Writer: file, hash, network, bytes.Buffer, io.Discard
```
- `io.Reader` / `io.Writer` are the `Stream` of Go, split into tiny single-method interfaces.
- `hash.Hash` implements `io.Writer`, so hashing a file is just `io.Copy(sha256.New(), f)`.
- `io.MultiWriter(a, b)` tees one stream into several writers, `io.TeeReader` does the same from the reading side, and `io.LimitReader` caps how much is read.
