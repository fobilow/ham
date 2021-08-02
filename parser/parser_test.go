package parser

import (
	"golang.org/x/net/html"
	"os"
	"testing"
)

func TestCompile(t *testing.T) {
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
