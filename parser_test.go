package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeTemplate writes the given content to a temp file in t.TempDir()
// and returns the path.
func writeTemplate(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "a.html")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// generate runs the Parser + Generator on the template at fpath and returns
// the bytes of the generated Go file, plus the block tree for direct
// inspection.
func generate(t *testing.T, fpath string) []byte {
	t.Helper()
	_, _, defFunc := (&Parser{}).Parse(fpath)
	if defFunc == nil {
		t.Fatal("no function definition found; template must contain a <%: func ... %> line")
	}

	dest := t.TempDir()
	g := Generator{PackageName: "template", Dest: dest}
	g.New(fpath)

	b, err := os.ReadFile(filepath.Join(dest, filepath.Base(fpath)+".go"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestMultipleBlocksStableOrdering is the regression test for the map-
// iteration nondeterminism in the old Parse() wiring loop. Two @-blocks
// in one template each carry trailing content in their After; the
// generated output must be byte-identical across many runs.
func TestMultipleBlocksStableOrdering(t *testing.T) {
	content := `<%: func T(buf *bytes.Buffer) %>
<%@ x { %>x-open<%@ } %>
<%@ y { %>y-open<%@ } %>
tail
`
	fpath := writeTemplate(t, content)
	first := generate(t, fpath)

	for i := 0; i < 20; i++ {
		got := generate(t, fpath)
		if !bytes.Equal(first, got) {
			t.Fatalf(
				"run %d produced different output than the first run\nfirst:\n%s\ngot:\n%s",
				i,
				first,
				got,
			)
		}
	}
}

// TestChildOverrideParentBlock checks the common inheritance case: a child
// template overrides two blocks defined in a plain-HTML layout (no func,
// no `~` in the layout itself) — a <meta> injection block and a <body>
// content block. The child's overrides must each appear exactly once in
// the generated output, the layout's defaults must not leak through, and
// the output must be stable across runs.
func TestChildOverrideParentBlock(t *testing.T) {
	dir := t.TempDir()
	layout := filepath.Join(dir, "layout.html")
	child := filepath.Join(dir, "page.html")

	// The layout is plain HTML with two blocks. No <%: func ... %> and no
	// <%~ ... %> — it just declares the structure and the two override
	// points (a meta slot for extra <meta> tags and a content slot for the
	// <body>). This is the typical "layout file" shape the user described.
	layoutHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <%@ meta { %><%@ } %>
</head>
<body>
    <%@ content { %><%@ } %>
</body>
</html>
`
	// The child overrides both blocks. It must carry a func definition so
	// the generator has a function to emit; the layout does not need one
	// because it is only ever pulled in via `~`.
	childHTML := `
<%: func Page(buf *bytes.Buffer) %>
<%~ "layout.html" %>
<%@ meta { %>
<meta name="description" content="child page">
<%@ } %>
<%@ content { %>
<h1>Child</h1>
<%@ } %>
`
	if err := os.WriteFile(layout, []byte(layoutHTML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(child, []byte(childHTML), 0644); err != nil {
		t.Fatal(err)
	}

	first := generate(t, child)

	// Both child overrides must be present exactly once.
	for _, want := range []string{`<meta name="description" content="child page">`, `<h1>Child</h1>`} {
		if c := bytes.Count(first, []byte(want)); c != 1 {
			t.Fatalf("%q appears %d times, want 1:\n%s", want, c, first)
		}
	}
	// The layout structure must be preserved around the overrides.
	for _, want := range []string{`<!DOCTYPE html>`, `<meta charset="utf-8">`, `<body>`, `</body>`} {
		if !bytes.Contains(first, []byte(want)) {
			t.Fatalf("layout structure %q missing from output:\n%s", want, first)
		}
	}
	// Ordering: meta block content must land inside <head> (before </head>),
	// content block must land inside <body>.
	metaIdx := bytes.Index(first, []byte(`name="description"`))
	headCloseIdx := bytes.Index(first, []byte("</head>"))
	bodyIdx := bytes.Index(first, []byte(`<h1>Child</h1>`))
	bodyCloseIdx := bytes.Index(first, []byte("</body>"))
	if metaIdx == -1 || headCloseIdx == -1 || bodyIdx == -1 || bodyCloseIdx == -1 ||
		!(metaIdx < headCloseIdx) || !(bodyIdx < bodyCloseIdx) || !(headCloseIdx < bodyIdx) {
		t.Fatalf(
			"blocks landed in the wrong part of the document:\nmetaIdx=%d headCloseIdx=%d bodyIdx=%d bodyCloseIdx=%d\n%s",
			metaIdx,
			headCloseIdx,
			bodyIdx,
			bodyCloseIdx,
			first,
		)
	}

	for i := 0; i < 20; i++ {
		got := generate(t, child)
		if !bytes.Equal(first, got) {
			t.Fatalf(
				"run %d produced different output than the first run\nfirst:\n%s\ngot:\n%s",
				i,
				first,
				got,
			)
		}
	}
}

// TestInclude checks that the `<%+ ... %>` directive splices the included
// file's blocks inline at the include's position in the tree. The included
// file may carry its own function definition (which the parent ignores, but
// the parent's rawCode is augmented with the child's rawCode).
func TestInclude(t *testing.T) {
	dir := t.TempDir()
	partial := filepath.Join(dir, "partial.html")
	main := filepath.Join(dir, "main.html")

	// A simple partial: some HTML, an escaped value, and a value.
	partialHTML := `<p class="partial">start</p>
<%= name %>
<p class="partial">end</p>
`

	// The main template defines the function, then includes the partial.
	mainHTML := `<%: func T(name string, buf *bytes.Buffer) %>
<%+ "partial.html" %>
<p class="tail">after</p>
`
	if err := os.WriteFile(partial, []byte(partialHTML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte(mainHTML), 0644); err != nil {
		t.Fatal(err)
	}

	first := generate(t, main)

	// The partial's content must be present, in order, between the include
	// site and the trailing <p class="tail">.
	partialStart := bytes.Index(first, []byte(`<p class="partial">start</p>`))
	escName := bytes.Index(first, []byte(`Goh.EscapeHTML(name, buf)`))
	partialEnd := bytes.Index(first, []byte(`<p class="partial">end</p>`))
	tail := bytes.Index(first, []byte(`<p class="tail">after</p>`))
	if partialStart == -1 || escName == -1 || partialEnd == -1 || tail == -1 ||
		!(partialStart < escName) || !(escName < partialEnd) || !(partialEnd < tail) {
		t.Fatalf(
			"included content missing or out of order:\npartialStart=%d escName=%d partialEnd=%d tail=%d\n%s",
			partialStart, escName, partialEnd, tail, first,
		)
	}

	for i := 0; i < 20; i++ {
		got := generate(t, main)
		if !bytes.Equal(first, got) {
			t.Fatalf(
				"run %d produced different output than the first run\nfirst:\n%s\ngot:\n%s",
				i, first, got,
			)
		}
	}
}

// TestParityScan pins the directive-pairing behavior that depends on the
// backtick/quote parity check in Parse(). A directive is only "closed" when
// the count of backticks and the count of double-quotes in its body are both
// even; otherwise the scanner keeps looking for the next "%>". These cases
// must keep working after the O(k^2) parity scan is replaced with an O(k)
// one.
func TestParityScan(t *testing.T) {
	bk := "`"  // a single backtick
	dq := "\"" // a single double-quote
	tests := []struct {
		name    string
		content string
		want    []string // substrings that must each appear exactly once
	}{
		{
			"simple-escaped-value",
			`<%: func T(buf *bytes.Buffer) %>
<%= name %>
`,
			[]string{"Goh.EscapeHTML(name, buf)"},
		},
		{
			"even-backticks-are-one-directive",
			`<%: func T(buf *bytes.Buffer) %>
<%! code := ` + bk + `a` + bk + ` %>
<%= name %>
`,
			[]string{"code := " + bk + "a" + bk, "Goh.EscapeHTML(name, buf)"},
		},
		{
			"odd-backticks-skip-one-close",
			// The body up to the first "%>" has an odd number of backticks (3),
			// so that "%>" is skipped and the *next* "%>" is the close. The
			// raw-code body captured is everything between `<%!` and the second
			// "%>", which still has 3 backticks. The trailing "tail" after the
			// real close must be emitted as HTML.
			`<%: func T(buf *bytes.Buffer) %>
<%! x := ` + bk + `a` + bk + `b` + bk + ` %> %>
tail
`,
			[]string{"tail"},
		},
		{
			"even-quotes-are-one-directive",
			`<%: func T(buf *bytes.Buffer) %>
<%! m := map[string]string{` + dq + `a` + dq + `: ` + dq + `b` + dq + `} %>
<%= name %>
`,
			[]string{"map[string]string{" + dq + "a" + dq + ": " + dq + "b" + dq + "}", "Goh.EscapeHTML(name, buf)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fpath := writeTemplate(t, tt.content)
			got := generate(t, fpath)
			for _, w := range tt.want {
				if c := bytes.Count(got, []byte(w)); c != 1 {
					t.Fatalf("%q appears %d times, want 1:\n%s", w, c, got)
				}
			}
		})
	}
}
