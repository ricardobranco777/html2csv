package htmltable

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestParseDirectoryListing_HappyPath_WithHRBoundary(t *testing.T) {
	src := `
<!doctype html><html><body>
<pre>
  <a href="?C=N;O=D">Name</a>
  <a href="?C=M;O=A">Last modified</a>
  <a href="?C=S;O=A">Size</a>
  <a href="?C=D;O=A">Description</a>
  <hr>
  <a href="/releases/amd64/">Parent Directory</a>                                                 -   
  <a href="file.iso">file.iso</a>                  2025-08-25 20:08  3.3G  
  <a href="file.iso.sha256">file.iso.sha256</a>          2025-08-25 20:08  112   
</pre>
</body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseDirectoryListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}

	if tab.Index != 1 || tab.Name != "directory" {
		t.Fatalf("unexpected table metadata: %+v", tab)
	}

	// Header must be extracted from <a> elements BEFORE <hr>, preserving multi-word header label.
	wantHeader := []string{"Name", "Last modified", "Size", "Description"}
	if len(tab.Rows) < 2 {
		t.Fatalf("expected at least header + 1 data row, got %d rows", len(tab.Rows))
	}
	assertSliceEqual(t, tab.Rows[0], wantHeader, "header")

	// Parent Directory row: name plus possibly "-" in size position; it has no date/time.
	// Implementation currently takes fields[0]+" "+fields[1] if available; here meta is "-" only,
	// so it will only produce []{"Parent Directory"} then normalize to header width.
	wantRow1 := []string{"Parent Directory", "", "", ""}
	assertSliceEqual(t, tab.Rows[1], wantRow1, "row[1]")

	// file.iso row: name, date/time, size, desc(empty)
	wantRow2 := []string{"file.iso", "2025-08-25 20:08", "3.3G", ""}
	assertSliceEqual(t, tab.Rows[2], wantRow2, "row[2]")

	wantRow3 := []string{"file.iso.sha256", "2025-08-25 20:08", "112", ""}
	assertSliceEqual(t, tab.Rows[3], wantRow3, "row[3]")
}

func TestParseDirectoryListing_IgnoresAnchorsBeforeHR_AsDataRows(t *testing.T) {
	// The header anchors must not become rows.
	src := `
<html><body><pre>
  <a href="?C=N;O=D">Name</a> <a href="?C=M;O=A">Last modified</a> <a href="?C=S;O=A">Size</a>
  <hr>
  <a href="x">x</a> 2025-01-01 00:00  1K
</pre></body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseDirectoryListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}

	if len(tab.Rows) != 2 {
		t.Fatalf("expected 2 rows (header + 1), got %d: %#v", len(tab.Rows), tab.Rows)
	}
	// Header should have 3 columns here (no Description anchor)
	assertSliceEqual(t, tab.Rows[0], []string{"Name", "Last modified", "Size"}, "header")
}

func TestParseDirectoryListing_NoPre_ReturnsFalse(t *testing.T) {
	src := `<html><body><div>no pre</div></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseDirectoryListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseDirectoryListing_NoHR_ReturnsFalse(t *testing.T) {
	// Without <hr>, everything is considered header => no rows.
	src := `<html><body><pre>
  <a href="?C=N;O=D">Name</a>
  <a href="?C=M;O=A">Last modified</a>
</pre></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseDirectoryListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseDirectoryListing_HasHRButNoHeaderAnchors_ReturnsFalse(t *testing.T) {
	// <hr> present but no <a> elements before it => no header.
	src := `<html><body><pre>
  some text
  <hr>
  <a href="x">x</a> 2025-01-01 00:00  1K
</pre></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseDirectoryListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseDirectoryListing_HasHeaderButNoRows_ReturnsFalse(t *testing.T) {
	src := `<html><body><pre>
  <a href="?C=N;O=D">Name</a>
  <a href="?C=S;O=A">Size</a>
  <hr>
</pre></body></html>`
	doc := mustParseHTML(t, src)

	_, ok := parseDirectoryListing(doc)
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestParseDirectoryListing_RowMetaNotTextNode_StillAddsRowNameAndNormalizes(t *testing.T) {
	// If the next sibling is not a TextNode, meta is empty and fields are empty.
	// Row should still exist with just the name, then normalized to header width.
	src := `<html><body><pre>
  <a href="?C=N;O=D">Name</a>
  <a href="?C=M;O=A">Last modified</a>
  <a href="?C=S;O=A">Size</a>
  <hr>
  <a href="x">x</a><span>ignored</span>
</pre></body></html>`

	doc := mustParseHTML(t, src)

	tab, ok := parseDirectoryListing(doc)
	if !ok {
		t.Fatalf("expected ok=true")
	}

	if len(tab.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d: %#v", len(tab.Rows), tab.Rows)
	}

	assertSliceEqual(t, tab.Rows[0], []string{"Name", "Last modified", "Size"}, "header")
	assertSliceEqual(t, tab.Rows[1], []string{"x", "", ""}, "row[1]")
}

func TestFirstElement_FindsFirstMatchingAtom(t *testing.T) {
	src := `<html><body>
  <div><pre id="first"></pre></div>
  <pre id="second"></pre>
</body></html>`
	doc := mustParseHTML(t, src)

	pre := firstElement(doc, atom.Pre)
	if pre == nil {
		t.Fatalf("expected to find <pre>")
	}

	// Ensure it's the first one in document order
	id := attr(pre, "id")
	if id != "first" {
		t.Fatalf("expected first pre id=first, got %q", id)
	}
}

func TestFirstElement_NoMatch_ReturnsNil(t *testing.T) {
	doc := mustParseHTML(t, `<html><body><div></div></body></html>`)
	if n := firstElement(doc, atom.Pre); n != nil {
		t.Fatalf("expected nil")
	}
}

// ---- helpers ----

func mustParseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}
	return doc
}

func assertSliceEqual(t *testing.T, got, want []string, label string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: len mismatch got=%d want=%d\ngot=%#v\nwant=%#v", label, len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: mismatch at %d got=%q want=%q\ngot=%#v\nwant=%#v", label, i, got[i], want[i], got, want)
		}
	}
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
