/* SPDX-License-Identifier: BSD-2-Clause */

package htmltable

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// parseLinkListing handles directory listings rendered as <ul>/<ol> lists of links.
// It emits a 1-column table: Name.
func parseLinkListing(doc *html.Node) (Table, bool) {
	// Prefer <ul>, fallback to <ol>.
	list := firstElement(doc, atom.Ul)
	if list == nil {
		list = firstElement(doc, atom.Ol)
	}
	if list == nil {
		return Table{}, false
	}

	var rows [][]string

	// Collect <a> inside immediate <li> children.
	for li := list.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.DataAtom != atom.Li {
			continue
		}

		a := firstChildElement(li, atom.A)
		if a == nil {
			continue
		}

		name := strings.TrimSpace(textContent(a))
		if name == "" {
			continue
		}

		rows = append(rows, []string{name})
	}

	if len(rows) <= 1 {
		return Table{}, false
	}

	return Table{
		Index: 1,
		Name:  "links",
		Rows:  rows,
	}, true
}

func firstChildElement(parent *html.Node, a atom.Atom) *html.Node {
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == a {
			return c
		}
	}
	return nil
}
