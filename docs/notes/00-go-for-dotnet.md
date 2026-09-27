# 00 — Go for a .NET developer

A map of the concepts you already know onto Go, and the places where the analogy breaks.

## Project structure

| .NET | Go | Notes |
|---|---|---|
| Solution (`.sln`) | Module (`go.mod`) | One module = one versioned unit, identified by its import path (`github.com/mpishchaev/music-archivist`). |
| Project / assembly | Package (a directory) | All `.go` files in one directory = one package. No project files. |
| Namespace | Package name | Import by path, refer by package name: `slog.Info(...)`. |
| NuGet `PackageReference` | `require` in `go.mod` | `go get pkg@version`, `go mod tidy` to clean up. `go.sum` ≈ lock file with hashes. |
| `internal` access modifier | `internal/` directory | Compiler-enforced: only code in the parent tree may import it. |
| `public` / `private` | Capitalized / lowercase name | `Track` is exported, `track` is package-private. Applies to types, funcs, fields, methods. |
| `static void Main` | `package main` + `func main()` | `os.Args`, `os.Exit(code)`. |

## Types

| .NET | Go |
|---|---|
| `class` | `struct` + methods. No inheritance. |
| `struct` (value type) | Every Go struct is a value type. Use `*T` when you want reference semantics. |
| Inheritance | Embedding: `type A struct{ B }` promotes `B`'s fields/methods. Composition, not "is-a". |
| `interface IFoo` + `: IFoo` | `type Foo interface{...}` — satisfied **implicitly** by any type with the methods. |
| Generics `List<T> where T : IComparable` | `func Max[T cmp.Ordered](a, b T) T` — type params with constraint interfaces. |
| `enum` | `type Kind int` + `const ( KindA Kind = iota; KindB )` |
| `null` | `nil` — only for pointers, interfaces, slices, maps, channels, funcs. Structs/ints have **zero values** instead (`0`, `""`, `false`). |
| Properties | Just exported fields. Getters only when there's logic; named `Name()`, not `GetName()`. |
| `List<T>` | slice `[]T` — `append`, `len`, `cap`, `s[i:j]`. Watch out: sub-slices share the backing array. |
| `Dictionary<K,V>` | `map[K]V` — `v, ok := m[k]`. Iteration order is random on purpose. |
| `HashSet<T>` | `map[T]struct{}` |

### Methods and receivers

```go
func (t Track) Label() string     { ... } // value receiver: gets a copy
func (t *Track) SetTitle(s string) { ... } // pointer receiver: can mutate, no copy
```
Rule of thumb: if any method needs a pointer receiver, make all of them pointer receivers.

## Errors instead of exceptions

| .NET | Go |
|---|---|
| `throw new X()` | `return fmt.Errorf("open %s: %w", path, err)` |
| `try/catch (FileNotFoundException)` | `if errors.Is(err, fs.ErrNotExist) { ... }` |
| `catch (MyException ex)` with data | `var e *MyError; if errors.As(err, &e) { ... }` |
| `InnerException` | `%w` wrapping, `errors.Unwrap` |
| `AggregateException` | `errors.Join(errs...)` |
| Crash on bug | `panic` — only for programmer errors, never for expected failures. |

Errors are values: `func Open(p string) (*File, error)`. Check them right away, and add context as the error travels up (`"scan: hash %s: %w"`).

## Resource cleanup

| .NET | Go |
|---|---|
| `using var f = File.Open(...)` | `f, err := os.Open(p); if err != nil {...}; defer f.Close()` |
| `finally` | `defer` — runs at **function** exit (not block exit), LIFO. |
| `IDisposable` | `io.Closer` (`Close() error`) — just an interface. |

## Concurrency

| .NET | Go |
|---|---|
| `Task.Run(...)` | `go f()` — starts a goroutine (cheap, ~KBs of stack). No return value; use channels or shared state. |
| `await` | There's no async/await. Code is written in blocking style, and the runtime parks goroutines on I/O. |
| `CancellationToken` | `context.Context` — always the first parameter: `func Do(ctx context.Context, ...)`. |
| `Channel<T>` / `BlockingCollection` | `chan T` — `ch <- v`, `v := <-ch`, `close(ch)`, `for v := range ch`. |
| `Task.WhenAll` | `sync.WaitGroup` or `errgroup.Group` (also cancels the rest on the first error). |
| `lock (obj)` | `sync.Mutex` — `mu.Lock(); defer mu.Unlock()` |
| `Parallel.ForEach` with degree | worker pool: N goroutines reading one channel, or `errgroup.SetLimit(n)`. |
| `Interlocked` | `sync/atomic` |

> "Don't communicate by sharing memory; share memory by communicating."

## Iteration

| .NET | Go |
|---|---|
| `foreach (var x in xs)` | `for i, x := range xs` (`_` to ignore the index) |
| `for (;;)` / `while` | `for { }` / `for cond { }`. `for` is the only loop keyword. |
| `IEnumerable<T>` + `yield return` | `iter.Seq[T]` — `func(yield func(T) bool)`, consumed with `for x := range seq`. |
| LINQ | No LINQ. Use plain loops + the `slices` and `maps` packages (`slices.SortFunc`, `slices.Contains`, `maps.Keys`). |

## Testing

| .NET (xUnit) | Go |
|---|---|
| test project | `foo_test.go` next to `foo.go` |
| `[Fact] public void X()` | `func TestX(t *testing.T)` |
| `[Theory] [InlineData]` | table-driven test: slice of cases + `t.Run(tc.name, ...)` |
| `Assert.Equal` | `assert.Equal(t, want, got)` (testify) or `if got != want { t.Errorf(...) }` |
| Moq | Hand-written fake that implements the small interface. |
| BenchmarkDotNet | `func BenchmarkX(b *testing.B)` + `go test -bench .` |
| — | `func FuzzX(f *testing.F)` — built-in fuzzing. |

## Tooling

| .NET | Go |
|---|---|
| `dotnet build` | `go build ./...` |
| `dotnet test` | `go test -race ./...` (`-race` = data race detector, always use it) |
| `dotnet format` / analyzers | `gofmt` (non-negotiable) + `go vet` + `golangci-lint` |
| `dotnet run` | `go run ./cmd/archivist` |

## Idioms to keep in mind

- **Accept interfaces, return structs.** Define the interface where it's *used*, and keep it small (1–3 methods).
- **Make the zero value useful.** `var b bytes.Buffer` is ready to use without a constructor.
- **Short names in small scopes** (`i`, `r`, `ctx`); descriptive names for exported identifiers.
- **Handle errors first, keep the happy path unindented** (`if err != nil { return ... }` early).
- **No getters/setters by default, no `I`-prefix on interfaces** (`Reader`, not `IReader`).
- **One way to format** — never argue about style, run the formatter.
