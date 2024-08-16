package ham

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	if err := c.compilePages(srcDir); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compilePages(dir string) error {
	publicDir := filepath.Join(c.workingDir, "public")
	pagesFiles, err := os.ReadDir(filepath.Join(c.workingDir, dir))
	if err != nil {
		return err
	}

	for _, page := range pagesFiles {
		// get file extension
		if filepath.Ext(page.Name()) != ".html" {
			log.Println("skipping file: " + page.Name())
			continue
		}

		pageName := page.Name()
		if page.IsDir() {
			if err := c.compilePages(filepath.Join(dir, pageName)); err != nil {
				return err
			}
			continue
		}

		fileName := filepath.Join(c.workingDir, dir, pageName)
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
			doc, hasEmbeds, err = c.compile(doc, fileName)
			if err != nil {
				return err
			}
			i++
		}

		// write final html to file
		pageFileName := filepath.Join(publicDir, pageName)
		log.Println("Creating page: " + pageFileName)
		if err := os.MkdirAll(filepath.Dir(pageFileName), os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(pageFileName, c.pageHTML, os.ModePerm); err != nil {
			return err
		}
		c.Reset()
	}
	return nil
}

func (c *Compiler) compile(doc *html.Node, pageFilePath string) (*html.Node, bool, error) {
	page := ParsePage(doc)

	pageCssFileName := strings.ReplaceAll(pageFilePath, ".html", ".css")
	page.Layout.CSS = append(page.Layout.CSS, pageCssFileName)

	pageTsFileName := strings.ReplaceAll(pageFilePath, ".html", ".ts")
	page.Layout.JsMod = append(page.Layout.JsMod, pageTsFileName)

	layoutFilePath := filepath.Join(filepath.Dir(pageFilePath), page.Layout.Src)
	if _, err := os.Stat(layoutFilePath); err != nil {
		return nil, false, fmt.Errorf("failed to compile %s. Layout file %s not found", pageFilePath, layoutFilePath)
	}
	log.Printf("Compiling Page: %s with %s\n", pageFilePath, layoutFilePath, page.Layout.CSS, page.Layout.Js, page.Layout.JsMod)

	buf := &bytes.Buffer{}
	if err := html.Render(buf, doc); err != nil {
		return nil, false, err
	}

	c.pageHTML = make([]byte, buf.Len())
	copy(c.pageHTML, buf.Bytes())

	if c.layoutHTML == nil {
		c.layoutHTML = readFile(layoutFilePath)
		doc, err := html.Parse(bytes.NewBuffer(c.layoutHTML))
		if err != nil {
			return nil, false, err
		}

		layout := ParseLayout(doc)
		buf.Reset()
		if err := html.Render(buf, doc); err != nil {
			return nil, false, err
		}

		dedupe := make(map[string]bool)
		c.layoutHTML = buf.Bytes()

		pageResources := append([]string{}, page.Layout.CSS...)
		pageResources = append(pageResources, page.Layout.Js...)
		pageResources = append(pageResources, page.Layout.JsMod...)

		pageCSS := make([]string, len(page.Layout.CSS))
		pageJs := make([]string, len(page.Layout.Js))
		for _, res := range pageResources {
			if !filepath.IsAbs(res) {
				res = filepath.Join(filepath.Dir(pageFilePath), res) // re-adjust res path
			}
			if _, ok := dedupe[res]; ok {
				continue
			}
			dedupe[res] = true
			if err := createFile(res, nil, false); err != nil {
				log.Println("error writing css file", err.Error())
			}
			i := strings.Index(res, filepath.Clean("/assets")) // make path os portable
			if i >= 0 {
				res = res[i:]
			} else {
				res = filepath.Join("assets", filepath.Base(res))
			}

			switch filepath.Ext(res) {
			case ".css":
				pageCSS = append(pageCSS, `<link rel="stylesheet" href="`+res+`">`)
			case ".js":
				pageJs = append(pageJs, `<script src="`+res+`"></script>`)
			case ".ts":
				res = strings.Replace(res, ".ts", ".js", 1)
				pageJs = append(pageJs, `<script type="module" src="`+res+`"></script>`)
			}
		}

		c.pageHTML = bytes.Replace(c.pageHTML, []byte("<html><head></head><body>"), []byte(""), 1) // strip out <html><head></head><body>
		c.pageHTML = bytes.Replace(c.pageHTML, []byte("</body></html>"), []byte(""), 1)            // strip out </html><body>
		c.pageHTML = bytes.Replace(c.layoutHTML, []byte("{ham:page}"), c.pageHTML, 1)
		c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte("{ham:css}"), []byte(strings.Join(pageCSS, "\n")))
		c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte("{ham:js}"), []byte(strings.Join(pageJs, "\n")))

		// find and replace layout embeds
		for _, embed := range layout.Embeds {
			if embed.Src != "" {
				embedFilePath := filepath.Join(filepath.Dir(layoutFilePath), embed.Src)
				log.Println("embedding", embedFilePath)
				embedContent := readFile(embedFilePath)
				c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte(embedPlaceholder(embed.Src)), embedContent)
			}
		}
	}

	// find and replace page embeds
	for _, embed := range page.Embeds {
		if embed.Src != "" {
			embedFilePath := filepath.Join(filepath.Dir(pageFilePath), embed.Src)
			log.Println("embedding", embedFilePath)
			embedContent := readFile(embedFilePath)
			c.pageHTML = bytes.ReplaceAll(c.pageHTML, []byte(embedPlaceholder(embed.Src)), embedContent)
		}
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

func createFile(filePath string, content []byte, override bool) error {
	if !override {
		if _, err := os.Stat(filePath); err == nil {
			return nil
		}
	}
	log.Println("Creating file: " + filePath)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filePath, content, os.ModePerm)
}
