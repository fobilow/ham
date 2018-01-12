package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/skratchdot/open-golang/open"
)

var outputDir string
var configFile string

func init(){
	configFile = filepath.Join(workingDir, "ham.json")
}

//TODO managing page content
//TODO inline editing
//TODO figure out how HAM will interact with other template
func main() {
	if len(os.Args) < 2 {
		fmt.Println("please provide a valid argument")
		os.Exit(1)
	}else{
		switch os.Args[1] {
		case "new":
			if len(os.Args) == 3 {
				//TODO validate site name (ensure it contains valid characters)
				newSite(os.Args[2])
			}else{
				fmt.Println("please provide a name for your site")
			}
			break
		case "build":
			//TODO add support for minification --no-minify
			if len(os.Args) == 3 {
				outputDir = os.Args[2]
				compile("./")
			}else{
				fmt.Println("please provide an output directory")
			}
			break
		case "serve":
			outputDir = "./hamed"
			compile("./")

			open.Start(fmt.Sprintf("%s:%d", "http://localhost", 4120))

			server(outputDir, "127.0.0.1", 4120)



			break
		}

	}
}