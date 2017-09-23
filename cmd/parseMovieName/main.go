package main

import (
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/scritch007/go-moviename"
)

func main() {
	logger := log.New("cmdLine")
	logger.SetLevel(log.DEBUG)
	mp := moviename.NewMovieParser(logger)
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s filename", os.Args[0])
		os.Exit(1)
	}
	m, err := mp.Parse(os.Args[1])
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%s (%d) => %s\n", m.Name, m.Year, m.Quality)
}
