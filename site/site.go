package ham

import (
	"fmt"
	"github.com/fobilow/ham/compiler"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"os"
	"path/filepath"
)

const configFileName = "ham.json"
const defaultCompileJSON = `{}`
const defaultLayout = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>HAM</title>
  <link type="ham/layout-css" href="../assets/css" />
</head>
<body>

<div id="app-info"></div>
<div class="container">
  <div class="row">
      
  </div>
</div>
<embed type="ham/layout-js" src="../assets/js" />
</body>
</html>
`
const defaultPage = `<div class="page"
     data-ham-layout='../layouts/default.html'
     data-ham-layout-css='[
     "../assets/css/css1.css",
     "../assets/css/css2.css"
     ]'
     data-ham-layout-js='[
     "../assets/js/js1.js",
     "../assets/js/js2.js"
     ]'
>
  <h1>Welcome to HAM</h1>
</div>`

var siteStructure = []string{
	"assets/css",
	"assets/js",
	"assets/img",
	"pages",
	"partials",
	"layouts",
}

type Site struct {
	host string
	port int
}

func New() *Site {
	return &Site{
		host: "localhost",
		port: 4120,
	}
}

func (h *Site) NewProject(siteName, workingDir string) error {
	// generate folder structure
	for _, folder := range siteStructure {
		if err := os.MkdirAll(filepath.Join(workingDir, siteName, folder), 0744); err != nil {
			return err
		}
	}

	// write default layout
	if err := os.WriteFile(filepath.Join(workingDir, siteName, "layouts", "default.html"), []byte(defaultLayout), os.ModePerm); err != nil {
		return err
	}

	// write default index.html
	if err := os.WriteFile(filepath.Join(workingDir, siteName, "pages", "index.html"), []byte(defaultPage), os.ModePerm); err != nil {
		return err
	}

	// write default ham.json
	return os.WriteFile(filepath.Join(workingDir, siteName, configFileName), []byte(defaultCompileJSON), os.ModePerm)
}

func (h *Site) Build(workingDir, outputDir string) error {
	c, err := compiler.New(workingDir, outputDir)
	if err != nil {
		return err
	}
	return c.Compile()
}

func (h *Site) Serve(workingDir string) error {
	outputDir := filepath.Join(workingDir, "hamed")
	c, err := compiler.New(workingDir, outputDir)
	if err != nil {
		return err
	}
	if err := c.Compile(); err != nil {
		return err
	}

	if err := open.Start(fmt.Sprintf("http://%s:%d", h.host, h.port)); err != nil {
		return err
	}

	absDocRoot, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	fmt.Println("Serving " + absDocRoot)
	http.Handle("/", http.FileServer(http.Dir(absDocRoot)))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", h.host, h.port), nil)
}

func (h *Site) Help() string {
	return `usage: ham <command> [<options>]

The following are supported HAM commands:
	new		Creates a new HAM site
	build		Compiles HAM site into html website
	serve		Starts a web server and serves a HAM site
	version		Displays version of HAM that you are running
`
}
