package main

import (
	"fmt"
	"os"
	"errors"
	"strings"
	"path/filepath"
	"regexp"
)

var Version string
var invalidCommandError = errors.New("please provide a valid action [new|build|serve|version|help]")

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

	if len(workingDirArg) > 0 {
		if !strings.HasPrefix(workingDirArg, "/"){
			Ham.workingDir = filepath.Join(Ham.workingDir, workingDirArg)
		}else{
			Ham.workingDir = workingDirArg
		}
	}

	Ham.version = Version

	switch os.Args[1] {
	case "new":
		Ham.NewSite(os.Args[2])
		break
	case "build":
		Ham.Build(os.Args[2])
		break
	case "serve":
		err := Ham.Serve()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Server started!")
		}
		break
	case "version":
		fmt.Println(Ham.Version())
		break
	case "help":
		fmt.Println(Ham.Help())
	}
}

func validateArgs() error {

	if len(os.Args) < 2 {
		return invalidCommandError
	}

	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			return errors.New("please provide a name for your site")
		}

		matched, err := regexp.MatchString("\\W+", os.Args[2])
		checkError(err)
		if matched {
			return errors.New("invalid project name. project name can only contains letters, digits or underscore")
		}
		break
	case "build":
		if len(os.Args) < 3 {
			return errors.New("please provide an output directory")
		}
		break
	default:
		return invalidCommandError
		break
	}

	return nil
}
