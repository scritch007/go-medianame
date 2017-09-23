package moviename

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

//Movie stores the movie Information
type Movie struct {
	Name        string
	Year        int
	Quality     int
	ProperCount int
}

//MovieParser configure Parser
type MovieParser struct {
	Logger log.Logger
}

//NewMovieParser create new movie parser
func NewMovieParser() MovieParser {
	return MovieParser{}
}

//Parse parse name and return Movie or error
func (m *MovieParser) Parse(name string) (Movie, error) {
	//Remove extension
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]
	fmt.Printf("Extension %s %s\n", ext, name)
	//Move [ ] to the end
	re := regexp.MustCompile("^\\[(?P<info>[^\\]]*)\\](?P<name>.*)")
	if re.MatchString(name) {
		tmp := re.FindStringSubmatch(name)

		name = fmt.Sprintf("%s %s", tmp[2], tmp[1])
	}

	//Replace special characters
	for _, c := range "[]()_,." {
		name = strings.Replace(name, string(c), " ", -1)
	}

	if strings.Index(name, " ") == -1 {
		name = strings.Replace(name, "-", " ", -1)
	}

	for _, c := range []string{"imax"} {
		name = strings.Replace(name, c, "", -1)
	}

	//Remove duplicated spaces
	spaceRe := regexp.MustCompile("  +")
	name = string(spaceRe.ReplaceAll([]byte(strings.TrimSpace(name)), []byte(" ")))

	return Movie{Name: name}, nil
}
