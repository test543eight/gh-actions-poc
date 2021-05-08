package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("vim-go")
	for _, e := range os.Environ() {
		fmt.Printf("%v.\n", e)
	}
}
