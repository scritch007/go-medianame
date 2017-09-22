package main

import (
	"fmt"
	"os"

	"github.com/scritch007/go-moviename"
)

func main() {
	mp := moviename.NewMovieParser()
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s filename", os.Args[0])
		os.Exit(1)
	}
	m, err := mp.Parse(os.Args[1])
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", m.Name)
}
