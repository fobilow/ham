package main

import (
	"net/http"
	"fmt"
	"path/filepath"
)

func init(){

}


func server(docRoot string, host string, port int) {
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

