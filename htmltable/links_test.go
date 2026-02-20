/* SPDX-License-Identifier: BSD-2-Clause */

package htmltable

import (
	"testing"
)

func TestParseLinkListing_UL_Simple(t *testing.T) {
	src := `
<html>
 <head>
  <title>Index of /ftp/pub/MidnightBSD/releases/amd64/ISO-IMAGES/4.0.2</title>
 </head>
 <body>
<h1>Index of /ftp/pub/MidnightBSD/releases/amd64/ISO-IMAGES/4.0.2</h1>
<ul>
  <li><a href="/ftp/pub/MidnightBSD/releases/amd64/ISO-IMAGES/"> Parent Directory</a></li>
  <li><a href="CHECKSUM.SHA256"> CHECKSUM.SHA256</a></li>
  <li><a href="CHECKSUM.SHA512"> CHECKSUM.SHA512</a></li>
  <li><a href="MidnightBSD-4.0.2--amd64-bootonly.iso"> MidnightBSD-4.0.2--amd64-bootonly.iso</a></li>
  <li><a href="MidnightBSD-4.0.2--amd64-disc1.iso"> MidnightBSD-4.0.2--amd64-disc1.iso</a></li>
  <li><a href="MidnightBSD-4.0.2--amd64-memstick.img"> MidnightBSD-4.0.2--amd64-memstick.img</a></li>
  <li><a href="MidnightBSD-4.0.2--amd64-mini-memstick.img"> MidnightBSD-4.0.2--amd64-mini-memstick.img</a></li>
</ul>
</body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseLinkListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if tab.Index != 1 || tab.Name != "links" {
		t.Fatalf("unexpected table metadata: %+v", tab)
	}

	want := [][]string{
		{"Parent Directory"},
		{"CHECKSUM.SHA256"},
		{"CHECKSUM.SHA512"},
		{"MidnightBSD-4.0.2--amd64-bootonly.iso"},
		{"MidnightBSD-4.0.2--amd64-disc1.iso"},
		{"MidnightBSD-4.0.2--amd64-memstick.img"},
		{"MidnightBSD-4.0.2--amd64-mini-memstick.img"},
	}

	assertRowsEqual(t, tab.Rows, want, "Rows")
}

func TestParseLinkListing_OL_Simple(t *testing.T) {
	src := `
<html><body>
<ol>
  <li><a href="a"> a </a></li>
  <li><a href="b">b</a></li>
</ol>
</body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseLinkListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}

	want := [][]string{
		{"a"},
		{"b"},
	}
	assertRowsEqual(t, tab.Rows, want, "Rows")
}

func TestParseLinkListing_NoList_ReturnsFalse(t *testing.T) {
	src := `<html><body><div><a href="x">x</a></div></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseLinkListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseLinkListing_ListWithoutAnchors_ReturnsFalse(t *testing.T) {
	src := `<html><body><ul><li>no link</li><li><span>x</span></li></ul></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseLinkListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseLinkListing_IgnoresNonLIChildren(t *testing.T) {
	src := `
<html><body>
<ul>
  text
  <li><a href="a">a</a></li>
  <!-- comment -->
  <li><a href="b"> b </a></li>
</ul>
</body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseLinkListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}

	want := [][]string{
		{"a"},
		{"b"},
	}
	assertRowsEqual(t, tab.Rows, want, "Rows")
}
