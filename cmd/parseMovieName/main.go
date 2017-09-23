package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/scritch007/go-moviename"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "v", false, "Enable debug logs")
}

func main() {
	flag.Parse()
	logger := log.New("cmdLine")
	if debug {
		logger.SetLevel(log.DEBUG)
	}
	mp := moviename.NewMovieParser(logger)
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s filename", os.Args[0])
		os.Exit(1)
	}
	m, err := mp.Parse(flag.Args()[0])
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%s (%d) => %s\n", m.Name, m.Year, m.Quality)
}
