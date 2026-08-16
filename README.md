> This project has been vibed coded, using Nemotron 3 Ultra, in about 15 minutes.

# vox

CLI dictionary for managing word definitions and reference links.

> vox – Latin for "voice" or "word"; very fast to type.

## Installation

```bash
go install github.com/yourusername/vox@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vox
cd vox
go build
```

## Usage

### Add a word

```bash
vox add <word> -d "definition" [-r "reference_url"] [-e "example"] [-l "title|url|description"]
```

Example:
```bash
vox add serendipity -d "the occurrence of events by chance in a happy way" \
  -r "https://en.wikipedia.org/wiki/Serendipity" \
  -e "Finding the book was pure serendipity" \
  -l "Wikipedia|https://en.wikipedia.org/wiki/Serendipity|Reference"
```

### Query a word (exact match)

```bash
vox query <word>
```

### Fuzzy search

```bash
vox fuzzy <pattern>
```

Finds words containing the pattern as a subsequence (e.g., `ap` matches `apple`, `application`, `map`).

### List all words

```bash
vox list
```

### Delete a word

```bash
vox delete <word>
```

### Export to JSON

```bash
vox export <output.json>
```

Output format:
```json
{
  "version": "1.0.0",
  "exported_at": "2026-08-16T...",
  "words": [...]
}
```

### Import from JSON

```bash
vox import <input.json> [--override]
```

- `--override`: Replace existing entries (default: skip existing)

## Data Storage

Data is stored in `~/.local/share/vox/words.json` (or `$XDG_DATA_HOME/vox/words.json` if set).
