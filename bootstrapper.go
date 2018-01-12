package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

)

var defaultCompileJson string
var siteStructure []string
var defaultLayout string
var defaultPage string
var installDir = "./"

func init(){
	defaultCompileJson = `{
  "index.html": {
    "layout": "default",
    "css": [
    ],
    "js": [
    ]
  }
}`
	defaultLayout = `<!DOCTYPE html>
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
	defaultPage = `<div class="page">
  <h1>Welcome to HAM</h1>
</div>`

	siteStructure = []string{
		"assets/css",
		"assets/js",
		"assets/img",
		"pages",
		"partials",
		"layouts",
	}

}


func newSite(siteName string) {


	//generate folder structure
	for _, folder := range siteStructure {
		//create output directory
		err := os.MkdirAll(filepath.Join(installDir, siteName, folder), 0744)
		checkError(err)
	}


	//write default layout
	ioutil.WriteFile(filepath.Join(installDir, siteName, "layouts", "default.html"), []byte(defaultLayout), os.ModePerm)

	//write default index.html
	ioutil.WriteFile(filepath.Join(installDir, siteName, "pages", "index.html"), []byte(defaultPage), os.ModePerm)

	//write default ham.json
	ioutil.WriteFile(filepath.Join(installDir, siteName, "ham.json"), []byte(defaultCompileJson), os.ModePerm)
}

