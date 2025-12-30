/* SPDX-License-Identifier: BSD-2-Clause */

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

import flag "github.com/spf13/pflag"

func main() {
	var delim string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVarP(&delim, "delimiter", "d", ",", "delimiter")
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

	if err := WriteTablesCSV(os.Stdout, tables, delimiter); err != nil {
		log.Fatal(err)
	}
}

func ExtractTables(doc *html.Node) [][][]string {
	var tables [][][]string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Table {
			rows := extractTableRows(n)
			if len(rows) > 0 {
				tables = append(tables, rows)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return tables
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

func WriteTablesCSV(w io.Writer, tables [][][]string, delim rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
	defer cw.Flush()

	for _, table := range tables {
		for _, row := range table {
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
