/* SPDX-License-Identifier: BSD-2-Clause */

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

import flag "github.com/spf13/pflag"

type Table struct {
	Index int
	ID    string
	Name  string
	Rows  [][]string
}

func main() {
	var delim, tableSel string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVarP(&delim, "delimiter", "d", ",", "delimiter")
	flag.StringVarP(&tableSel, "table", "t", "", "select tables by index or name")
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("ERROR: ")

	f := os.Stdin
	var err error

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() == 1 {
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	r := []rune(delim)
	if len(r) != 1 {
		fmt.Fprintf(os.Stderr, "delimiter must be a single character\n")
		os.Exit(1)
	}
	delimiter := r[0]

	doc, err := html.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	tables := ExtractTables(doc)

	sel := ParseTableSelector(tableSel)
	tables = FilterTables(tables, sel)

	if err := WriteTablesCSV(os.Stdout, tables, delimiter); err != nil {
		log.Fatal(err)
	}
}

func ExtractTables(doc *html.Node) []Table {
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

			rows := extractTableRows(n)
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

	return tables
}

func ParseTableSelector(s string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			m[p] = struct{}{}
		}
	}
	return m
}

func FilterTables(tables []Table, sel map[string]struct{}) []Table {
	if len(sel) == 0 {
		return tables
	}

	var out []Table
	for _, t := range tables {
		if _, ok := sel[strconv.Itoa(t.Index)]; ok {
			out = append(out, t)
			continue
		}
		if t.ID != "" {
			if _, ok := sel[t.ID]; ok {
				out = append(out, t)
				continue
			}
		}
		if t.Name != "" {
			if _, ok := sel[t.Name]; ok {
				out = append(out, t)
			}
		}
	}
	return out
}

func extractTableRows(table *html.Node) [][]string {
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

	normalizeRows(rows)
	return rows
}

func normalizeRows(rows [][]string) {
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

func WriteTablesCSV(w io.Writer, tables []Table, delim rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
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
