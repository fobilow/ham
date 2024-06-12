package ham

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	version := time.Now().Format("200602011504")
	if err := c.compilePages("pages", version); err != nil {
		return err
	}

	// if scripts directory exists, compile typescript files
	scriptsDir := filepath.Join(c.workingDir, "scripts")
	if _, err := os.Stat(scriptsDir); err == nil {
		log.Println("Compiling Scripts...")
		scriptFiles, err := os.ReadDir(scriptsDir)
		if err != nil {
			log.Println("scripts error", err.Error())
			return err
		}
		for _, scriptEntry := range scriptFiles {
			if scriptEntry.IsDir() {
				var tsConfigFile = filepath.Join(scriptsDir, scriptEntry.Name(), "tsconfig.json")
				if _, err := os.Stat(filepath.Join(tsConfigFile)); err == nil {
					log.Println("Typescript found. Compiling...")
					cmd := exec.Command("tsc")
					cmd.Dir = filepath.Join(scriptsDir, scriptEntry.Name())
					if out, err := cmd.CombinedOutput(); err != nil {
						log.Println("scripts error", string(out))
						return err
					}
				} else {
					log.Println("scripts error os.Stat", err.Error())
				}
			}
		}
	}

	// copy over assets
	assetsDir := filepath.Join(c.workingDir, "assets")
	if _, err := os.Stat(assetsDir); err == nil {
		src := assetsDir + "/"
		dest := filepath.Join(c.outputDir, "assets")
		cmd := exec.Command("cp", "-rf", src, dest)
		if _, err := cmd.CombinedOutput(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) compilePages(dir, version string) error {
	pagesDir := filepath.Join(c.workingDir, dir)
	pagesFiles, err := os.ReadDir(pagesDir)
	if err != nil {
		return err
	}

	for _, page := range pagesFiles {
		pageName := page.Name()
		if page.IsDir() {
			if err := c.compilePages(filepath.Join(dir, pageName), version); err != nil {
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
			doc, hasEmbeds, err = c.compile(doc, fileName, version)
			if err != nil {
				return err
			}
			i++
		}

		// write final html to file
		outDir := strings.Replace(dir, "pages", "", 1)
		pageFileName := filepath.Join(c.outputDir, outDir, pageName)
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

func (c *Compiler) compile(doc *html.Node, pageFilePath, version string) (*html.Node, bool, error) {
	page := ParsePage(doc)

	layoutFilePath := filepath.Join(filepath.Dir(pageFilePath), page.Layout.Src)
	log.Println("Compiling Page: " + pageFilePath + " with " + layoutFilePath)

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

		layout := ParseLayout(doc)
		buf.Reset()
		if err := html.Render(buf, doc); err != nil {
			return nil, false, err
		}

		c.layoutHTML = buf.Bytes()
		pageCSS := make([]string, len(page.Layout.CSS))
		for _, css := range page.Layout.CSS {
			css = filepath.Join(filepath.Dir(pageFilePath), css) // re-adjust css path
			if err := createFile(css, nil); err != nil {
				log.Println("error writing css file", err.Error())
			}
			css = css[strings.Index(css, "/assets"):]
			pageCSS = append(pageCSS, `<link rel="stylesheet" href="`+css+`?v=`+version+`">`)
		}

		pageJs := make([]string, len(page.Layout.Js))
		for _, js := range page.Layout.Js {
			js = filepath.Join(filepath.Dir(pageFilePath), js) // re-adjust js path
			if err := createFile(js, nil); err != nil {
				log.Println("error writing js file", err.Error())
			}
			js = js[strings.Index(js, "/assets"):]
			pageJs = append(pageJs, `<script src="`+js+`?v=`+version+`"></script>`)
		}
		for _, js := range page.Layout.JsMod {
			js = filepath.Join(filepath.Dir(pageFilePath), js) // re-adjust js path
			if err := createFile(js, nil); err != nil {
				log.Println("error writing js file", err.Error())
			}
			js = js[strings.Index(js, "/assets"):]
			pageJs = append(pageJs, `<script type="module" src="`+js+`?v=`+version+`"></script>`)
		}

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

func createFile(filePath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filePath, content, os.ModePerm)
}
