/* SPDX-License-Identifier: BSD-2-Clause */

package htmltable

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Table struct {
	Index int
	ID    string
	Name  string
	Rows  [][]string
}

func Parse(r io.Reader) ([]Table, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var tables []Table
	index := 0

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Table {
			index++

			var id, name string
			for _, a := range n.Attr {
				switch a.Key {
				case "id":
					id = a.Val
				case "name":
					name = a.Val
				}
			}

			rows := extractRows(n)
			if len(rows) > 0 {
				tables = append(tables, Table{
					Index: index,
					ID:    id,
					Name:  name,
					Rows:  rows,
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if len(tables) == 0 {
		if t, ok := parseDirectoryListing(doc); ok {
			tables = append(tables, t)
		}
	}

	return tables, nil
}

type Selector struct {
	Indexes map[int]struct{}
	Names   map[string]struct{}
}

func ParseSelector(s string) (Selector, error) {
	sel := Selector{
		Indexes: make(map[int]struct{}),
		Names:   make(map[string]struct{}),
	}

	if strings.TrimSpace(s) == "" {
		return sel, nil
	}

	for part := range strings.SplitSeq(s, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		if i, err := strconv.Atoi(p); err == nil {
			if i <= 0 {
				return sel, errors.New("table index must be >= 1")
			}
			sel.Indexes[i] = struct{}{}
		} else {
			sel.Names[p] = struct{}{}
		}
	}

	return sel, nil
}

func (s Selector) Apply(tables []Table) []Table {
	if len(s.Indexes) == 0 && len(s.Names) == 0 {
		return tables
	}

	var out []Table
	for _, t := range tables {
		if _, ok := s.Indexes[t.Index]; ok {
			out = append(out, t)
			continue
		}
		if _, ok := s.Names[t.ID]; ok {
			out = append(out, t)
			continue
		}
		if _, ok := s.Names[t.Name]; ok {
			out = append(out, t)
		}
	}
	return out
}

func SkipHeader(tables []Table) []Table {
	out := make([]Table, 0, len(tables))
	for _, t := range tables {
		if len(t.Rows) > 1 {
			t.Rows = t.Rows[1:]
			out = append(out, t)
		}
	}
	return out
}

type CSVEncoder struct {
	Comma rune
}

func NewCSVEncoder() *CSVEncoder {
	return &CSVEncoder{Comma: ','}
}

func (e *CSVEncoder) Encode(w io.Writer, tables []Table) error {
	cw := csv.NewWriter(w)
	cw.Comma = e.Comma
	defer cw.Flush()

	for _, t := range tables {
		for _, row := range t.Rows {
			if err := cw.Write(row); err != nil {
				return err
			}
		}
		_ = cw.Write([]string{})
	}
	return cw.Error()
}

func extractRows(table *html.Node) [][]string {
	var rows [][]string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Tr {
			var row []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.DataAtom == atom.Td || c.DataAtom == atom.Th) {
					row = append(row, strings.TrimSpace(textContent(c)))
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)

	normalize(rows)
	return rows
}

func normalize(rows [][]string) {
	maxCols := 0
	for _, r := range rows {
		if len(r) > maxCols {
			maxCols = len(r)
		}
	}
	for i := range rows {
		for len(rows[i]) < maxCols {
			rows[i] = append(rows[i], "")
		}
	}
}

func textContent(n *html.Node) string {
	var b strings.Builder

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return b.String()
}
