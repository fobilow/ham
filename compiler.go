package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"errors"
)


const configFileName = "ham.json"

type projectConfig struct {
	Layout string
	Css    []string
	Js     []string
}

type Project struct {
	dir string
}

//check that working directory is valid (i.e it contains a ham.json)
func (p Project) isValid() bool {
	configFile := p.getConfigFile()
	_, err := os.Stat(configFile)
	if err != nil {
		return false
	}

	return true
}

func (p Project) getConfigFile() string {
	return filepath.Join(p.dir, configFileName)
}

type Compiler struct {
	project Project
	outputDir string
}

func NewCompiler(project Project, outputDir string) *Compiler {

	if !project.isValid() {
		err := errors.New(project.dir + " is not a valid HAM project")
		checkError(err)
	}

	return &Compiler{project: project, outputDir: outputDir}
}

func (c *Compiler) compile() {

	// get layouts content
	layouts := c.getLayoutsMap()

	//get all partials placeholder
	partialsMap := c.getPartialsMap()

	//create output directory
	err := os.MkdirAll(c.outputDir, 0744)
	checkError(err)

	//loop through every page and replace partial placeholders
	pagesDir := filepath.Join(c.project.dir, "pages")
	pagesFiles, err := ioutil.ReadDir(pagesDir)
	checkError(err)

	var compileJsonData []byte
	compileJsonData, err = ioutil.ReadFile(c.project.getConfigFile())
	checkError(err)

	compileInfo := make(map[string]projectConfig)
	err = json.Unmarshal(compileJsonData, &compileInfo)
	checkError(err)

	t := time.Now()
	version := t.Format("200602011504")
	for _, file := range pagesFiles {
		pageName := file.Name()
		fileName := filepath.Join(pagesDir, pageName)

		layout := compileInfo[pageName].Layout
		layoutHtml := layouts[layout]

		fmt.Println("Compiling Page: " + fileName + " with " + layout)
		data, err := ioutil.ReadFile(fileName)
		checkError(err)

		pageHtml := string(data)

		pageCss := ""
		pageScript := ""

		pageHtml = strings.Replace(layoutHtml, "<page></page>", pageHtml, 1)
		pageHtml = strings.Replace(pageHtml, "{{v}}", version, 10)

		for _, css := range compileInfo[pageName].Css {
			pageCss += "<link rel=\"stylesheet\" href=\"" + css + "?v=" + version + "\">\n"
		}

		for _, js := range compileInfo[pageName].Js {
			pageScript += "<script src=\"" + js + "?v=" + version + "\"></script>\n"
		}

		pageHtml = strings.Replace(pageHtml, "<pagecss></pagecss>", pageCss, 1)
		pageHtml = strings.Replace(pageHtml, "<pagejs></pagejs>", pageScript, 1)

		complete := false

		for !complete {
			//replacing partials tags with partial html
			for tag, html := range partialsMap {
				replacer := strings.NewReplacer(tag, html)
				pageHtml = replacer.Replace(pageHtml)
			}

			foundUncompiledPartial := false
			for tag := range partialsMap {
				if strings.Contains(pageHtml, tag) {
					foundUncompiledPartial = true
					break
				}
			}

			if !foundUncompiledPartial {
				complete = true
			}
		}

		//write final html to file
		pageFileName := filepath.Join(c.outputDir, pageName)
		fmt.Println("Creating page: "+pageFileName)
		err = ioutil.WriteFile(pageFileName, []byte(pageHtml), os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	//copy over assets
	//TODO add support for minification --no-minify
	//TODO optmize, perhaps using rsync?
	cmd := exec.Command("cp", "-rf", filepath.Join(c.project.dir, "assets/"), filepath.Join(c.outputDir, "assets"))
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func (c *Compiler) getLayoutsMap() map[string]string {
	layouts := make(map[string]string)
	layoutDir := filepath.Join(c.project.dir, "layouts")
	layoutFiles, err := ioutil.ReadDir(layoutDir)
	checkError(err)
	for _, file := range layoutFiles {
		fmt.Println("Registering Layout: " + file.Name())

		data, err := ioutil.ReadFile(filepath.Join(layoutDir, file.Name()))
		checkError(err)

		b := strings.Replace(file.Name(), ".html", "", -1)

		layouts[b] = string(data)
	}

	return layouts
}

func (c *Compiler) getPartialsMap() map[string]string {

	replaceMap := make(map[string]string)

	partialsDir := filepath.Join(c.project.dir, "partials")
	partialFiles, err := ioutil.ReadDir(partialsDir)
	checkError(err)

	for _, file := range partialFiles {
		fn := strings.Replace(file.Name(), ".html", "", 1)
		tag := fmt.Sprintf("<%s></%s>", fn, fn)
		//tag := fmt.Sprintf("<div data-ham-partial=\"%s\"></div>", fn)
		data, err := ioutil.ReadFile(filepath.Join(partialsDir, file.Name()))
		checkError(err)

		replaceMap[tag] = string(data)
	}

	return replaceMap
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}


