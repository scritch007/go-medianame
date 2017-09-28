package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/labstack/gommon/log"
	"github.com/scritch007/go-medianame"
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
	mp := medianame.NewSerieParser(logger)
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s filename", os.Args[0])
		os.Exit(1)
	}
	m, err := mp.Parse(path.Base(flag.Args()[0]))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%s (%d) => %s\n", m.Name, m.Episode, m.Quality)
}
