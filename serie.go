package medianame

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

var (
	separators = "[/ -]"

	unwantedRegexps = []*regexp.Regexp{
		regexp.MustCompile("(?i)(\\d{1,3})\\s?x\\s?(0+)[^1-9]"), //5x0
		regexp.MustCompile("(?i)S(\\d{1,3})D(\\d{1,3})"),        //S3D1
		regexp.MustCompile("(?i)(?:s|series|\\b)\\s?\\d\\s?(?:&\\s?\\d)?[\\s-]*(?:complete|full)"),
		regexp.MustCompile("(?i)disc\\s\\d"),
	}

	//Make sure none of these are found embedded within a word or other numbers
	dateRegexps = []*regexp.Regexp{
		notInWord(fmt.Sprintf("(?i)(\\d{2,4})%s(\\d{1,2})%s(\\d{1,2})", separators, separators)),
		notInWord(fmt.Sprintf("(?i)(\\d{1,2})%s(\\d{1,2})%s(\\d{2,4})", separators, separators)),
		notInWord(fmt.Sprintf("(?i)(\\d{4})x(\\d{1,2})%s(\\d{1,2})", separators)),
		notInWord(fmt.Sprintf("(?i)(\\d{1,2})(?:st|nd|rd|th)?%s([a-z]{3,10})%s(\\d{4})", separators, separators)),
	}

	romanNumeralRe = "X{0,3}(?:IX|XI{0,4}|VI{0,4}|IV|V|I{1,4})"

	seasonPackRegexps = []*regexp.Regexp{
		//S01 or Season 1 but not Season 1 Episode|Part 2
		regexp.MustCompile(fmt.Sprintf("(?i)(?:season\\s?|s)(\\d{1,})(?:\\s|$)(?:^(?:(?:.*?\\s)?(?:episode|e|ep|part|pt)\\s?(?:\\d{1,3}|%s)|(?:\\d{1,3})\\s?of\\s?(?:\\d{1,3})))", romanNumeralRe)),
		regexp.MustCompile("(?i)(\\d{1,3})\\s?x\\s?all'"), // 1xAll
	}

	englishNumbers = []string{"one", "two", "three", "four", "five", "six", "seven",
		"eight", "nine", "ten"}

	epRegexps = []*regexp.Regexp{
		notInWord(fmt.Sprintf("(?:series|season|s)\\s?(\\d{1,4})(?:\\s(?:.*\\s)?)?(?:episode|ep|e|part|pt)\\s?(\\d{1,3}|%s)(?:\\s?e?(\\d{1,2}))?", romanNumeralRe)),
		notInWord(fmt.Sprintf("(?:series|season)\\s?(\\d{1,4})\\s(\\d{1,3})\\s?of\\s?(?:\\d{1,3})")),
		notInWord(fmt.Sprintf("(\\d{1,2})\\s?x\\s?(\\d+)(?:\\s(\\d{1,2}))?")),
		notInWord(fmt.Sprintf("(\\d{1,3})\\s?of\\s?(?:\\d{1,3})")),
		notInWord(fmt.Sprintf("(?:episode|e|ep|part|pt)\\s?(\\d{1,3}|%s)", romanNumeralRe)),
		notInWord(fmt.Sprintf("part\\s(%s)", strings.Join(englishNumbers, "|"))),
	}

	ignorePrefixes = []string{
		"(?:\\[[^\\[\\]]*\\])",
		"(?:HD.720p?:)",
		"(?:HD.1080p?:)",
		"(?:HD.2160p?:)",
	}
)

//Serie Represent serie object
type Serie struct {
	Name    string
	Episode int
	Season  int
	Quality string
}

//SerieParser parser object
type SerieParser struct {
	logger *log.Logger
}

//NewSerieParser create Parser
func NewSerieParser(logger *log.Logger) *SerieParser {
	return &SerieParser{
		logger: logger,
	}
}

func (s *SerieParser) guessName(name string) string {
	for _, c := range "_.,[]():" {
		name = strings.Replace(name, string(c), " ", -1)
	}
	matched, matchResult := s.parseIt(name, unwantedRegexps, dummyMatch)
	if matched {
		return ""
	}
	identifiedBy := ""
	matched, matchResult = s.parseIt(name, dateRegexps, dummyMatch)
	if matched {
		identifiedBy = "date"
	} else {
		matched, matchResult = s.parseIt(name, seasonPackRegexps, s.seasonCB)
		if !matched {
			matched, matchResult = s.parseIt(name, epRegexps, s.episodeCB)
		}
		identifiedBy = "ep"
	}
	if !matched {
		return ""
	}
	s.logger.Infof("Found a match %s", matchResult.Matches[0])
	if matchResult.Matches[0].Index > 1 {
		start := 0
		ignoreReg := regexp.MustCompile(strings.Join(ignorePrefixes, "|"))
		match := ignoreReg.FindString(name)
		if len(match) != 0 {
			start = strings.Index(name, match)
		}
		name = name[start:matchResult.Matches[0].Index]
		name = strings.Split(name, " - ")[0]
		specialReg := regexp.MustCompile("[\\._\\(\\) ]+")
		name = specialReg.ReplaceAllString(name, " ")
		name = strings.Trim(name, " -")
		name = strings.ToTitle(name)
		return name
	}
	s.logger.Debugf("Identified by %s", identifiedBy)
	return name
}

type matchCB func(matches []string) (bool, interface{})

func dummyMatch(matches []string) (bool, interface{}) {
	return true, nil
}

type match struct {
	Value string
	Index int
}
type matchResult struct {
	Matches []match
	context interface{}
}

func (s *SerieParser) parseIt(name string, regexps []*regexp.Regexp, cb matchCB) (bool, matchResult) {
	name = strings.ToLower(name)
	for _, re := range regexps {
		matches := re.FindAllString(name, -1)
		if len(matches) >= 1 {
			log.Infof("Found matches %v, %v", re, name)

			if matched, context := cb(matches); matched {
				res := matchResult{
					Matches: make([]match, len(matches)),
					context: context,
				}
				offset := 0
				for i, m := range matches {
					offset += strings.Index(name[offset:], m)
					res.Matches[i] = match{
						Value: m,
						Index: offset,
					}
				}
				return true, res
			}
		} else {
			s.logger.Debugf("No match for %s %s", re, name)
		}
	}
	return false, matchResult{}
}

func (s *SerieParser) seasonCB(matches []string) (bool, interface{}) {
	if len(matches) == 1 {
		return true, nil
	}
	return false, nil
}

type episodeMatch struct {
	Episode int
	Season  int
}

func (s *SerieParser) episodeCB(matches []string) (bool, interface{}) {
	season := 0
	episode := 0
	s.logger.Debugf("%v", matches)
	if len(matches) != 0 {
		var epError error
		strEp := ""
		if len(matches) == 2 {
			season, _ = strconv.Atoi(matches[0])
			episode, epError = strconv.Atoi(matches[1])
			strEp = matches[1]
		} else if len(matches) == 1 {
			season = 1
			episode, epError = strconv.Atoi(matches[0])
			strEp = matches[0]
		} else {
			s.logger.Errorf("Unknown matches length %d", len(matches))
			return false, nil
		}
		if epError != nil {
			//Let's convert it into int
			for i, num := range englishNumbers {
				if strEp == num {
					episode = i + 1
					epError = nil
					break
				}
			}
			if epError != nil {
				episode, epError = s.romanToInt(strEp)
			}
		}
		if epError != nil {
			s.logger.Errorf("Error retrieving information %v", epError)
			return false, nil
		}
		return true, &episodeMatch{
			Episode: episode,
			Season:  season,
		}
	}
	return false, nil
}

func (s *SerieParser) romanToInt(strEp string) (int, error) {
	//TODO
	return 0, errors.New("Couldn't find Value")
}

//Parse file name and return matching serie
func (s *SerieParser) Parse(name string) (Serie, error) {
	//Remove extension
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]
	name = s.guessName(name)
	return Serie{Name: name}, nil
}
