package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
	"fmt"
	"net/http"
)

type Ham struct {
	defaultCompileJson string
	siteStructure      []string
	defaultLayout      string
	defaultPage        string
	workingDir         string

	compiler *Compiler
}

func (h *Ham) Init(){

	h.defaultCompileJson = `{
  "index.html": {
    "layout": "default",
    "css": [
    ],
    "js": [
    ]
  }
}`
	h.defaultLayout = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>HAM</title>
  <pagecss></pagecss>
</head>
<body>

<div id="app-info"></div>
<div class="container">
  <div class="row">
      <page></page>
  </div>
</div>
<pagejs></pagejs>
</body>
</html>
`
	h.defaultPage = `<div class="page">
  <h1>Welcome to HAM</h1>
</div>`

	h.siteStructure = []string{
		"assets/css",
		"assets/js",
		"assets/img",
		"pages",
		"partials",
		"layouts",
	}

}

func NewHam() *Ham {
	h := &Ham{}

	workingDir, err := os.Getwd()
	checkError(err)

	h.workingDir = workingDir

	return h
}

func (h *Ham) NewSite(siteName string) {

	h.Init()

	//generate folder structure
	for _, folder := range h.siteStructure {
		//create output directory
		err := os.MkdirAll(filepath.Join(h.workingDir, siteName, folder), 0744)
		checkError(err)
	}

	//write default layout
	ioutil.WriteFile(filepath.Join(h.workingDir, siteName, "layouts", "default.html"), []byte(h.defaultLayout), os.ModePerm)

	//write default index.html
	ioutil.WriteFile(filepath.Join(h.workingDir, siteName, "pages", "index.html"), []byte(h.defaultPage), os.ModePerm)

	//write default ham.json
	ioutil.WriteFile(filepath.Join(h.workingDir, siteName, configFileName), []byte(h.defaultCompileJson), os.ModePerm)
}

func (h *Ham) Build(outputDir string) {

	project := Project{"", h.workingDir}
	h.compiler = NewCompiler(project, outputDir)

	h.compiler.compile()
}


func (h *Ham) Serve() {

	outputDir := filepath.Join(h.workingDir, "hamed")

	project := Project{"", h.workingDir}
	h.compiler = NewCompiler(project, outputDir)
	h.compiler.compile()

	port := 4120

	open.Start(fmt.Sprintf("%s:%d", "http://localhost", port))

	h.server(outputDir, "127.0.0.1", port)
}

func (h * Ham) server(docRoot string, host string, port int) {
	x, _ := filepath.Abs(docRoot)
	fmt.Println("Serving "+ x)
	http.Handle("/", http.FileServer(http.Dir(x)))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil{
		fmt.Println(err)
	}else{
		fmt.Println("Server started!")
	}
}

func (h *Ham) Version() {

}

func (h *Ham) Help() {


}