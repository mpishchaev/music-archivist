# music-archivist

CLI that scans scattered MP3 folders, indexes them into SQLite, detects duplicates,
recovers missing metadata and COPIES tracks into a clean library laid out by a template.

## Purpose of this repo — READ FIRST
This is a **learning project**. The user is a senior .NET developer (8+ yrs) refreshing Go.
Claude's job is to **teach**, not to ship code.

### Division of labor
- **Claude writes:** package layout, types, interfaces, function signatures, failing
  table-driven tests, `// TODO(you):` hints, `// LEARN:` notes, docs/notes/*.md.
- **User writes:** function bodies to make tests green.
- **Claude reviews:** points out non-idiomatic code, explains *why*, suggests the Go way.
  Do NOT silently rewrite the user's code — explain, then let the user fix it
  (unless explicitly asked "just fix it").
- Exception: boring glue (cobra wiring, SQL migrations, test fixtures) — Claude may write fully,
  marking it `// scaffold:`.
- Before each milestone: short briefing in chat — what Go concepts it exercises, .NET analogies.
- Never hand out the full solution for a TODO unless asked; escalate hints (nudge → pointer to
  stdlib func → snippet).

### Conventions for teaching material
- `// LEARN:` — 1–3 line inline note on an idiom, with a .NET analogy when helpful
  (e.g. `// LEARN: defer ~ finally/using; runs LIFO at function exit`).
- `// TODO(you):` — task for the user, with acceptance criteria / which test covers it.
- `docs/notes/NN-topic.md` — longer write-ups (errors, interfaces, concurrency, testing,
  database/sql, context, generics/iterators…), each with a ".NET ↔ Go" table.
- Code comments, godoc, notes: **English**. Chat with the user: **Russian**.

## Functional scope
- Input: one or more source roots (read-only, never modified).
- Only `.mp3` is processed. Other files are listed in the report as "skipped", untouched.
- Output: target root; path from a `text/template` in config, e.g.
  `{{.Artist}}/{{.Year}} - {{.Album}}/{{printf "%02d" .Track}} - {{.Title}}.mp3`.
  Unresolvable metadata → `_Unsorted/<original relative path>`.
- Operation: **copy only**. `--dry-run` prints the plan without touching the disk.
  Never overwrite an existing target file; collisions are reported.
- Duplicates:
  1. exact — SHA-256 of file content;
  2. logical — normalized (artist, title) + duration within tolerance;
  3. (later) acoustic — Chromaprint via `fpcalc`.
  From each group keep the "best" (bitrate, then has-full-tags, then size); others are reported, not copied.
- Metadata fallback chain: ID3 tags → filename/path parser → (later) MusicBrainz → `_Unsorted`.

## CLI (cobra)
- `archivist scan  --src DIR [--src DIR...]`  — walk, hash, read tags → SQLite index (incremental: skip unchanged by size+mtime).
- `archivist dupes`                            — print duplicate groups.
- `archivist plan  --dst DIR`                  — compute target paths, store/print plan.
- `archivist apply [--dry-run]`                — execute plan (copy).
- Global: `--config archivist.yaml`, `--db archivist.db`, `--log-level`, `--workers N`.

## Architecture
```
cmd/archivist/main.go        # entry; builds root cobra cmd, signal.NotifyContext
internal/cli/                # cobra commands; thin — parse flags, call services
internal/config/             # YAML config struct, defaults, Validate()
internal/model/              # Track, DuplicateGroup, PlanItem — plain structs
internal/scanner/            # walk fs.FS, filter .mp3, hash with worker pool (errgroup + ctx)
internal/meta/               # TagReader interface; id3 impl (dhowden/tag); filename parser
internal/index/              # SQLite repository (database/sql), migrations via embed.FS
internal/dedup/              # grouping strategies + "pick best"
internal/layout/             # text/template path render + filename sanitizing
internal/organizer/          # plan + apply (copy via io.Copy, atomic temp+rename)
internal/musicbrainz/        # (milestone 7) HTTP client, rate limit, httptest
internal/fingerprint/        # (milestone 8) os/exec fpcalc wrapper
testdata/                    # tiny MP3 fixtures
docs/notes/                  # learning notes
```
Design rules:
- Dependencies point inward: `cli → services → model`; services depend on small interfaces
  defined **at the consumer** (Go style, unlike .NET where interfaces live next to impl).
- "Accept interfaces, return structs". Constructors `NewX(deps...) *X`, no DI container.
- Every blocking/IO func takes `ctx context.Context` as first param.
- Errors: wrap with `fmt.Errorf("op: %w", err)`; sentinel errors (`ErrNoTags`) + `errors.Is/As`;
  no panics in library code; per-file errors are collected & reported, not fatal.
- Filesystem access via `fs.FS` where reading (testable with `fstest.MapFS`);
  writing via a small `FileSystem` interface or real OS in `t.TempDir()`.
- Logging: `log/slog`, logger passed explicitly (no globals).
- No global mutable state; no `init()` side effects.

## Stack
- Go 1.27, modules. `github.com/spf13/cobra`, `github.com/stretchr/testify` (assert/require),
  `github.com/dhowden/tag`, `modernc.org/sqlite` (pure Go), `gopkg.in/yaml.v3`,
  `golang.org/x/sync/errgroup`. Stdlib for everything else.
- Adding any new dependency → ask the user first and explain why stdlib isn't enough.

## Testing
- Table-driven tests with `t.Run`, `t.Parallel()` where safe.
- `testify/require` for preconditions, `assert` for checks.
- Fakes (hand-written structs implementing interfaces) over mocking frameworks.
- `t.TempDir()`, `fstest.MapFS`, in-memory SQLite (`file::memory:?cache=shared`), `httptest.Server`.
- Race detector always on. Fuzz test for filename parser; benchmark for hashing.

## Commands
- `go test -race ./...`
- `go vet ./...`
- `golangci-lint run` (config `.golangci.yml`: govet, staticcheck, errcheck, revive, gofumpt)
- `go run ./cmd/archivist scan --src ~/Music --dry-run`

## Milestones
0. Setup: go.mod, layout, Makefile/Taskfile, golangci-lint, first `// LEARN:` tour.
1. Scanner: walk + filter + SHA-256 with worker pool, ctx cancellation. (goroutines, channels, errgroup)
2. Metadata: TagReader interface + id3 impl; `ErrNoTags`. (interfaces, errors)
3. Index: SQLite schema, repository, incremental scan. (database/sql, embed, transactions)
4. Layout + plan + apply (copy, dry-run, collisions). (text/template, io, os)
5. Exact dedup. (maps, generics helper `GroupBy[K comparable, V any]`)
6. Filename/path parser fallback + fuzz test. (regexp, strings, fuzzing)
7. Logical dedup + MP3 duration/bitrate. (sorting, `slices`/`maps` packages)
8. MusicBrainz lookup. (net/http, JSON, rate limiting, httptest)
9. Chromaprint via fpcalc. (os/exec)

## Safety
- Source roots are never written to. Code must open sources read-only.
- Never run `apply` without `--dry-run` against real music in Claude-initiated commands.
