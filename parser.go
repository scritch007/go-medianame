package moviename

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

//Movie stores the movie Information
type Movie struct {
	Name        string
	Year        int
	Quality     string
	ProperCount int
}

var (
	specials = []string{"special", "bonus", "extra", "omake", "ova"}
	editions = []string{"dc", "extended", "uncut", "remastered", "unrated", "theatrical", "chrono", "se"}
	cutoffs  = []string{"limited", "xvid", "h264", "x264", "h.264", "x.264", "screener", "unrated", "3d", "extended", "directors", "director\\'s", "multisubs", "dubbed", "subbed", "multi"}
	propers  = []string{"proper", "repack", "rerip", "real", "final"}
)

func init() {
	cutoffs = append(cutoffs, specials...)
	cutoffs = append(cutoffs, editions...)
}

//MovieParser configure Parser
type MovieParser struct {
	logger *log.Logger
}

//NewMovieParser create new movie parser
func NewMovieParser(logger *log.Logger) MovieParser {

	return MovieParser{logger: logger}
}

func isLetter(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func stringInSlice(s string, slice []string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}
	return false
}

//Parse parse name and return Movie or error
func (m *MovieParser) Parse(name string) (Movie, error) {
	//Remove extension
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]
	m.logger.Debugf("Extension %s %s\n", ext, name)
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

	parts := strings.Split(name, " ")
	cutPart := 256
	allCaps := true
	properCount := 0
	yearPos := 0
	year := 0

	for partPos, part := range parts {
		m.logger.Debugf("Currently Dealing with: %s => %d", part, partPos)
		cut := false
		if partPos < 1 {
			continue
		}
		num, err := strconv.Atoi(part)
		if err == nil {
			m.logger.Debugf("Found a number %d, %d", num, time.Now().Year())
			//Consider that this is a number
			if 1930 < num && num <= time.Now().Year() {
				m.logger.Debugf("Seems Like it's a year")
				if yearPos == cutPart {
					cutPart = partPos
				}
				year = num
				yearPos = partPos
				cut = true
			}
		}
		upper := strings.ToUpper(part)
		if upper != part {
			allCaps = false
		}
		if len(part) > 3 && upper == part && isLetter(part) && !allCaps {
			m.logger.Debugf("Yes cut me, I'm a letter and after not caps")
			cut = true
		}

		lower := strings.ToLower(part)
		if stringInSlice(lower, cutoffs) {
			m.logger.Debugf("Yes I'm a cutoff")
			cut = true
		}
		if stringInSlice(lower, propers) {
			m.logger.Debugf("Yes I'm a proper")
			cut = true
			if stringInSlice(lower, []string{"real", "final"}) || year != 0 {
				m.logger.Debugf("I'm not in the movie Name ")
				properCount++
				cut = true
			}
		}
		m.logger.Debugf("%b, %d %d", cut, partPos, cutPart)
		if cut && partPos < cutPart {
			cutPart = partPos
		}
	}
	if cutPart != 256 {
		m.logger.Infof("Parts: %v, cut is: %v\n", parts, parts[cutPart])
	}

	absCut := len(strings.Join(parts[:cutPart], " "))
	m.logger.Infof("after parts check, cut data would be: `%s` abs_cut: %d\n", name[:absCut], absCut)
	return Movie{Name: name[:absCut], Year: year, Quality: m.ParseQuality(name)}, nil
}
