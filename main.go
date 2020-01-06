package main

import (
	"fmt"
	"os"
	"errors"
	"strings"
	"path/filepath"
)

var Version string

func init() {
	Version = "1.0.0"
}

//TODO figure out how HAM will interact with other template
func main() {

	err := validateArgs()
	checkError(err)

	Ham := NewHam()
	var workingDirArg string
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-wd=") {
			workingDirArg = strings.Trim(arg[4:], "''\"\"")
			break
		}
	}

	var workingDir string
	workingDir, err = os.Getwd()
	checkError(err)

	if len(workingDirArg) > 0 && !strings.HasPrefix(workingDirArg, "/") {
		workingDir = filepath.Join(workingDir, workingDirArg)
	}

	Ham.workingDir = workingDir
	fmt.Println("Working Dir: " + workingDir)


	switch os.Args[1] {
	case "new":
		Ham.NewSite(os.Args[2] + ".ham")
		break
	case "build":
		//TODO add support for minification --no-minify
		Ham.Build(os.Args[2])
		break
	case "serve":
		Ham.Serve()
		break
	case "version":
		Ham.Version()
		break
	case "help":
		Ham.Help()
	}
}

func validateArgs() error {

	if len(os.Args) < 2 {
		return errors.New("please provide a valid argument")
	}

	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			return errors.New("please provide a name for your site")
		} else {
			//TODO validateArgs site name (ensure it contains valid characters)
		}
		break
	case "build":
		//TODO add support for minification --no-minify
		if len(os.Args) < 3 {
			return errors.New("please provide an output directory")
		}
		break
	case "serve":
		break
	case "version":
		break
	case "help":
	default:
		return errors.New("please provide a valid action [new|build|serve|version|help]")
		break
	}

	return nil
}
