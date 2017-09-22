package moviename

import (
	"fmt"
	"log"
	"regexp"
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
	//Move [ ] to the end
	re := regexp.MustCompile("^\\[(?P<info>.*)\\](?P<name>.*)")
	if re.MatchString(name) {
		reversed := fmt.Sprintf("${%s} ${%s}", re.SubexpNames()[2], re.SubexpNames()[1])

		name = re.ReplaceAllString(name, reversed)
	}
	return Movie{Name: name}, nil
}
