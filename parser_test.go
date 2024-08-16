package ham

import (
	"encoding/json"
	"os"
	"testing"

	"golang.org/x/net/html"
)

func TestParsePage(t *testing.T) {
	file, _ := os.Open("./test-site/src/index.html")

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
	file, _ := os.Open("./test-site/src/default.lhtml")

	// parse dom
	doc, _ := html.Parse(file)
	layout := ParseLayout(doc)

	want := 3
	got := len(layout.Embeds)
	if got != want {
		t.Errorf("parse failed: expected %d but got %d page embeds", want, got)
	}
}

func TestConfigParse(t *testing.T) {
	config := `{
     "layout": "../layouts/authed.html",
     "css": [
     "../assets/css/global.css"
     ],
     "js":[
     "../assets/js/app.js"
     ]}`

	var layout Layout
	err := json.Unmarshal([]byte(config), &layout)
	if err != nil {
		t.Errorf("parse failed: %v", err)
		return
	}

	if layout.Src != "../layouts/authed.html" {
		t.Errorf("parse failed: expected %s but got %s", "../layouts/authed.html", layout.Src)
	}
	if len(layout.CSS) != 1 {
		t.Errorf("parse failed: expected %d but got %d", 1, len(layout.CSS))
	}
	if len(layout.Js) != 1 {
		t.Errorf("parse failed: expected %d but got %d", 1, len(layout.Js))
	}
	if len(layout.JsMod) != 0 {
		t.Errorf("parse failed: expected %d but got %d", 0, len(layout.JsMod))
	}
}
