# music-archivist

CLI that scans scattered MP3 folders, indexes them into SQLite, detects duplicates,
recovers missing metadata and **copies** tracks into a clean library laid out by a template.

> Learning project: relearning Go through a practical tool. See `docs/notes/` for write-ups.

## Usage (planned)

```sh
archivist scan  --src ~/Music --src /mnt/old-disk/mp3
archivist dupes
archivist plan  --dst ~/Library/Music
archivist apply --dry-run
archivist apply
```

Sources are never modified: files are only copied.

## Development

```sh
make test   # go test -race ./...
make lint   # golangci-lint run
```
