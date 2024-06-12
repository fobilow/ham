package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fobilow/ham"
)

var Version string

func init() {
	Version = "1.0.0"
}

var validSiteName = regexp.MustCompile(`\W+`)

func main() {
	h := ham.NewSite()
	newCmd := newFlagSet(h, "init")
	buildCmd := newFlagSet(h, "build")
	serveCmd := newFlagSet(h, "serve")

	bwd := buildCmd.String("w", "./", "working directory")
	bod := buildCmd.String("o", "./"+ham.DefaultOutputDirName, "output directory")
	swd := serveCmd.String("w", "./", "working directory")

	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "init":
		name := os.Args[2]
		if len(name) == 0 {
			fmt.Println("please provide a name for your site")
			newCmd.Usage()
			return
		}
		if validSiteName.MatchString(name) {
			fmt.Println(name)
			fmt.Println("invalid project name. project name can only contains letters, digits or underscore")
			newCmd.Usage()
			return
		}
		checkError(h.NewProject(name, getWorkingDir("./")))
	case "build":
		checkError(buildCmd.Parse(os.Args[2:]))
		if len(*bwd) == 0 {
			fmt.Println("please provide a working directory")
			buildCmd.Usage()
			return
		}
		if len(*bod) == 0 {
			fmt.Println("please provide an output directory")
			buildCmd.Usage()
			return
		}
		checkError(h.Build(getWorkingDir(*bwd), *bod))
	case "serve":
		checkError(serveCmd.Parse(os.Args[2:]))
		if len(*swd) == 0 {
			fmt.Println("please provide a working directory")
			serveCmd.Usage()
			return
		}
		checkError(h.Serve(getWorkingDir(*swd)))
		fmt.Println("Server started!")
	case "version":
		fmt.Println("Version: " + Version)
	default:
		fmt.Println(h.Help())
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getWorkingDir(wd string) string {
	defaultWorkingDir, err := os.Getwd()
	checkError(err)

	workingDir := wd
	if !strings.HasPrefix(wd, "/") {
		workingDir = filepath.Join(defaultWorkingDir, wd)
	}
	return workingDir
}

func newFlagSet(s *ham.Site, name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = func() { fmt.Println(s.Help()) }
	return fs
}
