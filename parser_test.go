package ham

import (
	"os"
	"testing"

	"golang.org/x/net/html"
)

func TestParsePage(t *testing.T) {
	file, _ := os.Open("../test-site/pages/index.html")

	// parse dom
	doc, _ := html.Parse(file)
	page := ParsePage(doc)

	want := 4
	got := len(page.Embeds)
	if got != want {
		t.Errorf("parse failed: expected %d but got %d page embeds", want, got)
	}

	want = 2
	got = len(page.Layout.Js)
	if got != want {
		t.Errorf("parse failed: expected %d but got %d layout Js embeds", want, got)
	}

	want = 2
	got = len(page.Layout.CSS)
	if got != want {
		t.Errorf("parse failed: expected %d but got %d layout CSS embeds", want, got)
	}
}

func TestParseLayout(t *testing.T) {
	file, _ := os.Open("../test-site/layouts/default.html")

	// parse dom
	doc, _ := html.Parse(file)
	layout := ParseLayout(doc)

	want := 3
	got := len(layout.Embeds)
	if got != want {
		t.Errorf("parse failed: expected %d but got %d page embeds", want, got)
	}
}
