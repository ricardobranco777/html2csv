![Build Status](https://github.com/ricardobranco777/html2csv/actions/workflows/ci.yml/badge.svg)

# html2csv

Transform HTML tables to CSV.

## Usage

```
Usage: html2csv [OPTIONS] FILE
  -d, --delimiter string   delimiter (default ",")
  -t, --table string       select tables by index or name
```

## Notes

- If a file is not specified, read from stdin
- The delimiter must be a single character
