package parser

import (
	"encoding/json"
	"golang.org/x/net/html"
	"log"
)

type Layout struct {
	Src    string
	CSS    []string
	Js     []string
	Embeds []Embed
}

type Page struct {
	Layout *Layout
	Embeds []Embed
}

type Embed struct {
	Type string
	Src  string
}

func ParseLayout(doc *html.Node) Layout {
	var layout Layout
	parseLayout(doc, &layout)
	return layout
}

func parseLayout(start *html.Node, layout *Layout) {
	if layout == nil {
		panic("layout is nil")
	}
	// fmt.Println("layout transversing ", start.Data)
	if start.Type == html.ElementNode {
		switch start.Data {
		case "link":
			for _, attr := range start.Attr {
				if attr.Key == "type" && attr.Val == "ham/layout-css" {
					start.Data = "{ham:css}"
					start.Type = 1
					layout.Embeds = append(layout.Embeds, Embed{Type: attr.Val})
				}
			}
		case "embed":
			em := Embed{}
			for _, attr := range start.Attr {
				if attr.Key == "type" {
					em.Type = attr.Val
				}
				if attr.Key == "src" {
					em.Src = attr.Val
				}
			}
			layout.Embeds = append(layout.Embeds, em)
			start.Type = 1
			switch em.Type {
			case "ham/partial":
				start.Data = EmbedPlaceholder(em.Src)
			case "ham/page":
				start.Data = "{ham:page}"
			case "ham/layout-js":
				start.Data = "{ham:js}"
			}
		}
	}

	for start = start.FirstChild; start != nil; start = start.NextSibling {
		parseLayout(start, layout)
	}
}

func ParsePage(doc *html.Node) Page {
	var page Page
	parsePage(doc, &page)
	return page
}

func parsePage(start *html.Node, page *Page) {
	if page.Layout == nil {
		page.Layout = &Layout{}
	}
	// fmt.Println("page transversing ", start.Data)
	if start.Type == html.ElementNode {
		switch start.Data {
		case "embed":
			em := Embed{}
			for _, attr := range start.Attr {
				if attr.Key == "type" {
					em.Type = attr.Val
				}
				if attr.Key == "src" {
					em.Src = attr.Val
				}
			}
			page.Embeds = append(page.Embeds, em)
			// replace <embed> tag with placeholders
			start.Type = 1
			start.Data = EmbedPlaceholder(em.Src)
		case "div":
			var newAttr []html.Attribute
			for _, attr := range start.Attr {
				switch attr.Key {
				case "data-ham-layout":
					page.Layout.Src = attr.Val
				case "data-ham-layout-css":
					var css []string
					if err := json.Unmarshal([]byte(attr.Val), &css); err != nil {
						log.Println(err.Error())
						continue
					}
					page.Layout.CSS = css
				case "data-ham-layout-js":
					var js []string
					if err := json.Unmarshal([]byte(attr.Val), &js); err != nil {
						log.Println(err.Error())
						continue
					}
					page.Layout.Js = js
				default:
					newAttr = append(newAttr, attr)
				}
				start.Attr = newAttr
			}
		}
	}

	for start = start.FirstChild; start != nil; start = start.NextSibling {
		parsePage(start, page)
	}
}

func EmbedPlaceholder(src string) string {
	return "{embed:" + src + "}"
}
