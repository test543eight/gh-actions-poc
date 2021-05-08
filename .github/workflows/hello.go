package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fmt.Println("vim-go")
	for _, e := range os.Environ() {
		fmt.Printf("%v.\n", e)
	}
	b, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		fmt.Printf("--> %v.\n", err)
	}
	fmt.Printf("--> b= %v.\n", string(b))
}
