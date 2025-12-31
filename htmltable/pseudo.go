/* SPDX-License-Identifier: BSD-2-Clause */

package htmltable

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func parseDirectoryListing(doc *html.Node) (Table, bool) {
	pre := firstElement(doc, atom.Pre)
	if pre == nil {
		return Table{}, false
	}

	var header []string
	var rows [][]string
	inHeader := true

	for n := pre.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.DataAtom == atom.Hr {
			inHeader = false
			continue
		}

		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			text := strings.TrimSpace(textContent(n))

			if inHeader {
				header = append(header, text)
				continue
			}

			var meta string
			if s := n.NextSibling; s != nil && s.Type == html.TextNode {
				meta = strings.TrimSpace(s.Data)
			}

			fields := strings.Fields(meta)

			row := []string{text}
			if len(fields) >= 2 {
				row = append(row, fields[0]+" "+fields[1])
			}
			if len(fields) >= 3 {
				row = append(row, fields[2])
			}

			rows = append(rows, row)
		}
	}

	if len(header) == 0 || len(rows) == 0 {
		return Table{}, false
	}

	// Normalize rows to header width
	for i := range rows {
		for len(rows[i]) < len(header) {
			rows[i] = append(rows[i], "")
		}
	}

	return Table{
		Index: 1,
		Name:  "directory",
		Rows:  append([][]string{header}, rows...),
	}, true
}

func firstElement(root *html.Node, a atom.Atom) *html.Node {
	var found *html.Node
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if found != nil {
			return
		}
		if n.Type == html.ElementNode && n.DataAtom == a {
			found = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(root)
	return found
}
