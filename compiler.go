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



type comileConfig struct {
	Layout string
	Css    []string
	Js     []string
}

var workingDir string

func compile(wd string) {
	workingDir = wd

	//check that working directory is valid (i.e it contains a ham.json)
	_, err := os.Stat(configFile)
	if err != nil {
		err = errors.New(workingDir+ " is not a valid HAM project")
	}
	checkError(err)


	// get layouts content
	layouts := getLayoutsMap()

	//get all partials placeholder
	partialsMap := getPartialsMap()

	//create output directory
	err = os.MkdirAll(outputDir, 0744)
	checkError(err)

	//loop through every page and replace partial placeholders
	pagesDir := filepath.Join(workingDir, "pages")
	pagesFiles, err := ioutil.ReadDir(pagesDir)
	checkError(err)

	var compileJsonData []byte
	compileJsonData, err = ioutil.ReadFile(configFile)
	checkError(err)

	compileInfo := make(map[string]comileConfig)
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
		pageFileName := filepath.Join(outputDir, pageName)
		fmt.Println("Creating page: "+pageFileName)
		err = ioutil.WriteFile(pageFileName, []byte(pageHtml), os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
		}



	}

	//copy over assets
	cmd := exec.Command("cp", "-rf", filepath.Join(workingDir, "assets/"), filepath.Join(outputDir, "assets"))
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	//write version no
	versionStr := fmt.Sprintf("{\"%s\":\"%s\"}", "version", t.Format("2006.01.02-1504"))
	ioutil.WriteFile(filepath.Join(outputDir, "version.json"), []byte(versionStr), os.ModePerm)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getLayoutsMap() map[string]string {
	layouts := make(map[string]string)
	layoutDir := filepath.Join(workingDir, "layouts")
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

func getPartialsMap() map[string]string {

	replaceMap := make(map[string]string)

	partialsDir := filepath.Join(workingDir, "partials")
	partialFiles, err := ioutil.ReadDir(partialsDir)
	checkError(err)

	for _, file := range partialFiles {
		fn := strings.Replace(file.Name(), ".html", "", 1)
		tag := fmt.Sprintf("<%s></%s>", fn, fn)
		data, err := ioutil.ReadFile(filepath.Join(partialsDir, file.Name()))
		checkError(err)

		replaceMap[tag] = string(data)
	}

	return replaceMap
}
