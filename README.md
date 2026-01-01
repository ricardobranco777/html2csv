![Build Status](https://github.com/ricardobranco777/html2csv/actions/workflows/ci.yml/badge.svg)

# html2csv

Transform HTML tables to CSV.

Docker image available at `ghcr.io/ricardobranco777/html2csv:latest`

## Usage

```
Usage: html2csv [OPTIONS] FILE
  -d, --delimiter string   delimiter (default ",")
  -H, --no-header          skip table header
  -t, --table string       select tables by index or name
  -T, --tsv                use TAB as delimiter
      --version            print version and exit
```

## Notes

- If a file is not specified, read from stdin
- The delimiter must be a single character
