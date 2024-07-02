package ham

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const DefaultOutputDirName = "hamout"
const configFileName = "ham.json"
const defaultCompileJSON = `{}`
const defaultLayout = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>HAM</title>
  <link type="ham/layout-css"/>
</head>
<body>

<div id="app-info"></div>
<div class="container">
  <div class="row">
      <embed type="ham/page"/>
  </div>
</div>
<embed type="ham/layout-js"/>
</body>
</html>
`
const defaultPage = `<div class="page"
	data-ham-page-config='{
      "layout": "../layouts/default.html"
     }'
>
  <h1>Welcome to HAM</h1>
</div>`

const defaultTsConfig = `{
  "compilerOptions": {
    "target": "es2020",                                  /* Set the JavaScript language version for emitted JavaScript and include compatible library declarations. */
    "module": "node16",                                /* Specify what module code is generated. */
    "esModuleInterop": true,                             /* Emit additional JavaScript to ease support for importing CommonJS modules. This enables 'allowSyntheticDefaultImports' for type compatibility. */
    "forceConsistentCasingInFileNames": true,            /* Ensure that casing is correct in imports. */
    "strict": true,                                      /* Enable all strict type-checking options. */
    "skipLibCheck": true,                                 /* Skip type checking all .d.ts files. */
    "verbatimModuleSyntax": true,
    "outDir": "../assets/%s/js"
  }
}`
const tsconfigFileName = "tsconfig.json"
const defaultPackageJSON = `{
	  "name": "%s",
	  "version": "1.0.0",
	  "description": "",
	  "type": "module"
}`

var siteStructure = []string{
	"assets/{site-name}/css",
	"assets/{site-name}/js",
	"assets/{site-name}/img",
	"pages",
	"partials",
	"layouts",
	"scripts",
}

type Site struct {
	host string
	port int
}

func NewSite() *Site {
	return &Site{
		host: "localhost",
		port: 4120,
	}
}

func (h *Site) NewProject(siteName, workingDir string) error {
	// generate folder structure
	for _, folder := range siteStructure {
		folder = strings.ReplaceAll(folder, "{site-name}", siteName)
		if err := os.MkdirAll(filepath.Join(workingDir, siteName, folder), 0744); err != nil {
			return err
		}
	}

	// write default layout
	if err := createFile(filepath.Join(workingDir, siteName, "layouts", "default.html"), []byte(defaultLayout), false); err != nil {
		return err
	}

	// write default index.html
	if err := createFile(filepath.Join(workingDir, siteName, "pages", "index.html"), []byte(defaultPage), false); err != nil {
		return err
	}

	// write default tsconfig.json
	if err := createFile(filepath.Join(workingDir, siteName, tsconfigFileName), []byte(fmt.Sprintf(defaultTsConfig, siteName)), false); err != nil {
		return err
	}

	// write default package.json
	if err := createFile(filepath.Join(workingDir, siteName, "package.json"), []byte(fmt.Sprintf(defaultPackageJSON, siteName)), false); err != nil {
		return err
	}

	// write default ham.json
	return createFile(filepath.Join(workingDir, siteName, configFileName), []byte(defaultCompileJSON), true)
}

func (h *Site) Build(workingDir, outputDir string) error {
	c, err := New(workingDir, outputDir)
	if err != nil {
		return err
	}
	return c.Compile()
}

func (h *Site) Serve(workingDir string) error {
	outputDir := filepath.Join(workingDir, DefaultOutputDirName)
	if err := h.Build(workingDir, outputDir); err != nil {
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
  init		Creates a new HAM site
  build		Compiles HAM site into html website
  serve		Starts a web server and serves a HAM site
  version	Displays version of HAM that you are running
`
}
