package compiler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fobilow/ham/parser"
	"golang.org/x/net/html"
)

const parseLimit = 1000 // max number of times to iterate and find partials inside partials
type Compiler struct {
	workingDir string
	outputDir  string
	pageHTML   []byte
	layoutHTML []byte
}

func New(workingDir, outputDir string) (*Compiler, error) {
	if _, err := os.Stat(filepath.Join(workingDir, "ham.json")); err != nil {
		return nil, fmt.Errorf("%s  is not a valid HAM project", workingDir)
	}

	return &Compiler{workingDir: workingDir, outputDir: outputDir}, nil
}

func (c *Compiler) Compile() error {
	// create output directory
	if err := os.MkdirAll(c.outputDir, 0744); err != nil {
		return err
	}

	// loop through every page and replace partial placeholders
	pagesDir := filepath.Join(c.workingDir, "pages")
	pagesFiles, err := ioutil.ReadDir(pagesDir)
	if err != nil {
		return err
	}

	version := time.Now().Format("200602011504")
	for _, pageFile := range pagesFiles {
		pageName := pageFile.Name()
		fileName := filepath.Join(pagesDir, pageName)

		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		// parse dom
		doc, err := html.Parse(file)
		if err != nil {
			return err
		}

		hasEmbeds := true
		i := 0
		for hasEmbeds && i < parseLimit {
			doc, hasEmbeds, err = c.compilePage(doc, fileName, version)
			if err != nil {
				return err
			}
			i++
		}

		// write final html to file
		pageFileName := filepath.Join(c.outputDir, pageName)
		log.Println("Creating page: " + pageFileName)
		if err := os.MkdirAll(filepath.Dir(pageFileName), os.ModePerm); err != nil {
			return err
		}
		if err := ioutil.WriteFile(pageFileName, c.pageHTML, os.ModePerm); err != nil {
			return err
		}
		c.Reset()
	}

	// copy over assets
	src := filepath.Join(c.workingDir, "assets") + "/"
	dest := filepath.Join(c.outputDir, "assets")
	cmd := exec.Command("cp", "-rf", src, dest)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compilePage(doc *html.Node, fileName, version string) (*html.Node, bool, error) {
	page := parser.ParsePage(doc)
	layoutFilePath := c.resolvePath(page.Layout.Src)
	log.Println("Compiling Page: " + fileName + " with " + layoutFilePath)

	buf := &bytes.Buffer{}
	if err := html.Render(buf, doc); err != nil {
		return nil, false, err
	}

	c.pageHTML = make([]byte, buf.Len())
	copy(c.pageHTML, buf.Bytes())

	if c.layoutHTML == nil {
		c.pageHTML = bytes.Replace(c.pageHTML, []byte("<html><head></head><body>"), []byte(""), 1) // strip out <html><head></head><body>
		c.pageHTML = bytes.Replace(c.pageHTML, []byte("</body></html>"), []byte(""), 1)            // strip out </html><body>

		c.layoutHTML = readFile(layoutFilePath)
		doc, err := html.Parse(bytes.NewBuffer(c.layoutHTML))
		if err != nil {
			return nil, false, err
		}

		layout := parser.ParseLayout(doc)
		buf.Reset()
		if err := html.Render(buf, doc); err != nil {
			return nil, false, err
		}

		c.layoutHTML = buf.Bytes()
		pageCSS := make([]string, len(page.Layout.CSS))
		for _, css := range page.Layout.CSS {
			css = strings.Replace(css, "../", "", 1) // re-adjust css path
			pageCSS = append(pageCSS, `<link rel="stylesheet" href="`+css+`?v=`+version+`">`)
		}

		pageJs := make([]string, len(page.Layout.Js))
		for _, js := range page.Layout.Js {
			js = strings.Replace(js, "../", "", 1) // re-adjust js path
			pageJs = append(pageJs, `<script src="`+js+`?v=`+version+`"></script>`)
		}
		c.pageHTML = bytes.Replace(c.layoutHTML, []byte("{ham:page}"), c.pageHTML, 1)
		c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte("{ham:css}"), []byte(strings.Join(pageCSS, "\n")))
		c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte("{ham:js}"), []byte(strings.Join(pageJs, "\n")))

		// find and replace layout embeds
		for _, embed := range layout.Embeds {
			embedContent := readFile(c.resolvePath(embed.Src))
			c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte(parser.EmbedPlaceholder(embed.Src)), embedContent)
		}
	}

	// find and replace page embeds
	for _, embed := range page.Embeds {
		log.Println("embedding", embed.Src)
		embedContent := readFile(c.resolvePath(embed.Src))
		c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte(parser.EmbedPlaceholder(embed.Src)), embedContent)
	}

	doc, err := html.Parse(bytes.NewBuffer(c.pageHTML))
	if err != nil {
		return nil, false, err
	}

	return doc, len(page.Embeds) > 0, nil
}

func (c *Compiler) Reset() {
	c.pageHTML = nil
	c.layoutHTML = nil
}

var readCache map[string][]byte

func readFile(filename string) []byte {
	if readCache == nil {
		readCache = make(map[string][]byte)
	}

	if _, ok := readCache[filename]; !ok {
		file, err := os.ReadFile(filename)
		if err != nil {
			return nil
		}
		readCache[filename] = file
	}

	return readCache[filename]
}

func (c *Compiler) resolvePath(path string) string {
	return filepath.Join(c.workingDir, strings.Replace(path, "../", "", 1))
}
