package htmltable

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// ---- Parse() tests ----

func TestParse_ExtractsMultipleTables_WithIndexAndAttrs(t *testing.T) {
	src := `
<!doctype html><html><body>
  <table id="t1" name="alpha">
    <tr><th>A</th><th>B</th></tr>
    <tr><td>1</td><td>2</td></tr>
  </table>

  <div>
    <table id="t2">
      <tr><th>X</th><th>Y</th><th>Z</th></tr>
      <tr><td>p</td><td>q</td><td>r</td></tr>
    </table>
  </div>
</body></html>`

	tables, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}

	if tables[0].Index != 1 || tables[0].ID != "t1" || tables[0].Name != "alpha" {
		t.Fatalf("unexpected table[0] metadata: %+v", tables[0])
	}
	if tables[1].Index != 2 || tables[1].ID != "t2" || tables[1].Name != "" {
		t.Fatalf("unexpected table[1] metadata: %+v", tables[1])
	}

	want0 := [][]string{
		{"A", "B"},
		{"1", "2"},
	}
	want1 := [][]string{
		{"X", "Y", "Z"},
		{"p", "q", "r"},
	}
	assertRowsEqual(t, tables[0].Rows, want0, "table[0].Rows")
	assertRowsEqual(t, tables[1].Rows, want1, "table[1].Rows")
}

func TestParse_RealTable_TrimsEmptyIconColumn_DropsEmptyRows_NormalizesRagged(t *testing.T) {
	// Simulates directory-listing tables with an empty icon column and structural empty rows.
	// Also includes a ragged row (missing a cell) to ensure normalization pads.
	src := `
<!doctype html><html><body>
<table>
  <tr>
    <th></th><th>Name</th><th>Size</th><th>Date</th>
  </tr>

  <tr><td colspan="4"></td></tr> <!-- empty structural row -->

  <tr>
    <td></td><td>[parent directory]</td><td></td><td></td>
  </tr>

  <tr>
    <td></td><td>file1</td><td>10</td><td>2025-01-01</td>
  </tr>

  <tr>
    <td></td><td>ragged</td><td>99</td>
  </tr>
</table>
</body></html>`

	tables, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}

	// Expected:
	// - leading empty icon column removed
	// - empty structural row removed
	// - ragged row padded to 3 cols (Name, Size, Date)
	want := [][]string{
		{"Name", "Size", "Date"},
		{"[parent directory]", "", ""},
		{"file1", "10", "2025-01-01"},
		{"ragged", "99", ""},
	}

	assertRowsEqual(t, tables[0].Rows, want, "Rows")
}

func TestParse_FallsBackToPseudoOnlyWhenNoRealTables(t *testing.T) {
	// Document contains <pre> listing and a real table.
	// Parse should return only the table(s), not pseudo fallback.
	src := `
<!doctype html><html><body>

<pre>
  <a href="?C=N;O=D">Name</a> <a href="?C=M;O=A">Last modified</a> <a href="?C=S;O=A">Size</a> <a href="?C=D;O=A">Description</a>
  <hr>
  <a href="/x/">Parent Directory</a> -
</pre>

<table id="real">
  <tr><th>H</th></tr>
  <tr><td>v</td></tr>
</table>

</body></html>`

	tables, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected 1 table (real), got %d", len(tables))
	}
	if tables[0].ID != "real" {
		t.Fatalf("expected real table, got ID=%q", tables[0].ID)
	}
	assertRowsEqual(t, tables[0].Rows, [][]string{{"H"}, {"v"}}, "Rows")
}

func TestParse_ErrorFromReader(t *testing.T) {
	r := &errReader{err: errors.New("boom")}
	_, err := Parse(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type errReader struct{ err error }

func (e *errReader) Read(p []byte) (int, error) { return 0, e.err }

// ---- ParseSelector / Apply tests ----

func TestParseSelector_EmptyAndWhitespace(t *testing.T) {
	sel, err := ParseSelector("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sel.Indexes) != 0 || len(sel.Names) != 0 {
		t.Fatalf("expected empty selector, got %+v", sel)
	}
}

func TestParseSelector_MixedIndexesAndNames(t *testing.T) {
	sel, err := ParseSelector(" 1,foo,  2 ,bar,, ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := sel.Indexes[1]; !ok {
		t.Fatalf("expected index 1 selected")
	}
	if _, ok := sel.Indexes[2]; !ok {
		t.Fatalf("expected index 2 selected")
	}
	if _, ok := sel.Names["foo"]; !ok {
		t.Fatalf("expected name foo selected")
	}
	if _, ok := sel.Names["bar"]; !ok {
		t.Fatalf("expected name bar selected")
	}
}

func TestParseSelector_InvalidIndex(t *testing.T) {
	for _, in := range []string{"0", "-1", " 0,foo"} {
		_, err := ParseSelector(in)
		if err == nil {
			t.Fatalf("expected error for %q, got nil", in)
		}
	}
}

func TestSelectorApply_SelectsByIndexOrIDOrName(t *testing.T) {
	tables := []Table{
		{Index: 1, ID: "t1", Name: "alpha"},
		{Index: 2, ID: "t2", Name: "beta"},
		{Index: 3, ID: "", Name: "gamma"},
	}

	sel := Selector{
		Indexes: map[int]struct{}{2: {}},
		Names:   map[string]struct{}{"t1": {}, "gamma": {}},
	}
	got := sel.Apply(tables)

	// Expected: index 2, id t1, name gamma -> tables[1], tables[0], tables[2] in traversal order
	if len(got) != 3 {
		t.Fatalf("expected 3 tables, got %d", len(got))
	}
	if got[0].Index != 1 { // matched by ID t1
		t.Fatalf("expected first match Index=1, got %d", got[0].Index)
	}
	if got[1].Index != 2 { // matched by index 2
		t.Fatalf("expected second match Index=2, got %d", got[1].Index)
	}
	if got[2].Index != 3 { // matched by name gamma
		t.Fatalf("expected third match Index=3, got %d", got[2].Index)
	}
}

func TestSelectorApply_EmptySelectorReturnsInput(t *testing.T) {
	tables := []Table{{Index: 1}, {Index: 2}}
	sel := Selector{Indexes: map[int]struct{}{}, Names: map[string]struct{}{}}

	got := sel.Apply(tables)
	if len(got) != len(tables) {
		t.Fatalf("expected unchanged length, got %d", len(got))
	}
	// Same order, same elements (shallow check)
	for i := range tables {
		if got[i].Index != tables[i].Index {
			t.Fatalf("unexpected element at %d", i)
		}
	}
}

// ---- SkipHeader tests ----

func TestSkipHeader_DropsTablesWithOnlyHeader(t *testing.T) {
	in := []Table{
		{Index: 1, Rows: [][]string{{"h"}}},
		{Index: 2, Rows: [][]string{{"h"}, {"r1"}}},
		{Index: 3, Rows: [][]string{{"h"}, {"r1"}, {"r2"}}},
	}

	out := SkipHeader(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 tables after SkipHeader, got %d", len(out))
	}

	if len(out[0].Rows) != 1 || out[0].Rows[0][0] != "r1" {
		t.Fatalf("unexpected rows for table 2: %#v", out[0].Rows)
	}
	if len(out[1].Rows) != 2 || out[1].Rows[0][0] != "r1" || out[1].Rows[1][0] != "r2" {
		t.Fatalf("unexpected rows for table 3: %#v", out[1].Rows)
	}
}

// ---- CSVEncoder tests ----

func TestCSVEncoder_Encode_DefaultDelimiterAndBlankLineBetweenTables(t *testing.T) {
	tables := []Table{
		{Rows: [][]string{{"a", "b"}, {"c", "d"}}},
		{Rows: [][]string{{"1"}, {"2"}}},
	}

	var buf bytes.Buffer
	enc := NewCSVEncoder()
	if err := enc.Encode(&buf, tables); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Note: csv.Writer uses \n line endings.
	want := "a,b\nc,d\n\n1\n2\n\n"
	if buf.String() != want {
		t.Fatalf("unexpected CSV output:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestCSVEncoder_Encode_CustomDelimiter(t *testing.T) {
	tables := []Table{
		{Rows: [][]string{{"a", "b"}}},
	}

	var buf bytes.Buffer
	enc := NewCSVEncoder()
	enc.Comma = ';'
	if err := enc.Encode(&buf, tables); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	want := "a;b\n\n"
	if buf.String() != want {
		t.Fatalf("unexpected CSV output: %q want %q", buf.String(), want)
	}
}

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestCSVEncoder_Encode_PropagatesWriterError(t *testing.T) {
	tables := []Table{
		{Rows: [][]string{{"a", "b"}, {"c", "d"}}},
	}

	enc := NewCSVEncoder()
	err := enc.Encode(errWriter{}, tables)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---- Internal helper behavior tests ----

func TestTrimEmptyColumns_RaggedDoesNotPanicAndTrimsLeadingEmptyColumn(t *testing.T) {
	// Header has 2 cols, data row has only 1 col, and a leading empty col exists for all rows.
	rows := [][]string{
		{"", "H1", "H2"},
		{"", "x"}, // ragged
		{"", "y", "z"},
	}

	// Should not panic and should remove col 0.
	out := trimEmptyColumns(rows)

	want := [][]string{
		{"H1", "H2"},
		{"x", ""}, // col padded later by normalize in extractRows; here we only test trim
		{"y", "z"},
	}
	// Note: trimEmptyColumns pads missing kept columns with "".
	assertRowsEqual(t, out, want, "trimEmptyColumns output")
}

func TestDropEmptyRows_RemovesAllEmptyRows(t *testing.T) {
	rows := [][]string{
		{"a", ""},
		{"", ""},
		{" ", "\t"},
		{"b", "c"},
	}

	out := dropEmptyRows(rows)
	want := [][]string{
		{"a", ""},
		{"b", "c"},
	}
	assertRowsEqual(t, out, want, "dropEmptyRows output")
}

func TestNormalize_PadsToMaxWidth(t *testing.T) {
	rows := [][]string{
		{"a"},
		{"b", "c", "d"},
		{"e", "f"},
	}

	normalize(rows)

	want := [][]string{
		{"a", "", ""},
		{"b", "c", "d"},
		{"e", "f", ""},
	}
	assertRowsEqual(t, rows, want, "normalize output")
}

func TestTextContent_ConcatenatesNestedText(t *testing.T) {
	// Use Parse() to get a node from the same parser, then call textContent on a td node.
	src := `<html><body><table><tr><td>Hello <b>World</b>!</td></tr></table></body></html>`
	tables, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(tables) != 1 || len(tables[0].Rows) != 1 || len(tables[0].Rows[0]) != 1 {
		t.Fatalf("unexpected parsed shape: %+v", tables)
	}
	// This validates textContent indirectly via extractRows; expect "Hello World!"
	if tables[0].Rows[0][0] != "Hello World!" {
		t.Fatalf("expected %q got %q", "Hello World!", tables[0].Rows[0][0])
	}
}

// ---- test helpers ----

func assertRowsEqual(t *testing.T, got, want [][]string, label string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s: row count mismatch: got %d want %d\ngot=%#v\nwant=%#v", label, len(got), len(want), got, want)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("%s: row[%d] col count mismatch: got %d want %d\ngot=%#v\nwant=%#v",
				label, i, len(got[i]), len(want[i]), got[i], want[i])
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("%s: mismatch at [%d][%d]: got %q want %q\nrow got=%#v\nrow want=%#v",
					label, i, j, got[i][j], want[i][j], got[i], want[i])
			}
		}
	}
}
