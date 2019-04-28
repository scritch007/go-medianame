package medianame

import (
	"errors"
	"fmt"
	"path/filepath"

	"strconv"
	"strings"

	regexp "github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/labstack/gommon/log"
)

var (
	separators = "[/ -]"

	unwantedRegexps = []regexp.Regexp{
		regexp.MustCompile("(\\d{1,3})\\s?x\\s?(0+)[^1-9]", regexp.CASELESS), //5x0
		regexp.MustCompile("S(\\d{1,3})D(\\d{1,3})", regexp.CASELESS),        //S3D1
		regexp.MustCompile("(?:s|series|\\b)\\s?\\d\\s?(?:&\\s?\\d)?[\\s-]*(?:complete|full)", regexp.CASELESS),
		regexp.MustCompile("disc\\s\\d", regexp.CASELESS),
	}

	//Make sure none of these are found embedded within a word or other numbers
	dateRegexps = []regexp.Regexp{
		notInWord(fmt.Sprintf("(\\d{2,4})%s(\\d{1,2})%s(\\d{1,2})", separators, separators)),
		notInWord(fmt.Sprintf("(\\d{1,2})%s(\\d{1,2})%s(\\d{2,4})", separators, separators)),
		notInWord(fmt.Sprintf("(\\d{4})x(\\d{1,2})%s(\\d{1,2})", separators)),
		notInWord(fmt.Sprintf("(\\d{1,2})(?:st|nd|rd|th)?%s([a-z]{3,10})%s(\\d{4})", separators, separators)),
	}

	romanNumeralRe = "X{0,3}(?:IX|XI{0,4}|VI{0,4}|IV|V|I{1,4})"

	seasonPackRegexps = []regexp.Regexp{
		//S01 or Season 1 but not Season 1 Episode|Part 2
		regexp.MustCompile(fmt.Sprintf("(?:season\\s?|s)(\\d{1,})(?:\\s|$)(?!(?:(?:.*?\\s)?(?:episode|e|ep|part|pt)\\s?(?:\\d{1,3}|%s)|(?:\\d{1,3})\\s?of\\s?(?:\\d{1,3})))", romanNumeralRe), regexp.CASELESS),
		regexp.MustCompile("(\\d{1,3})\\s?x\\s?all'", regexp.CASELESS), // 1xAll
	}

	englishNumbers = []string{"one", "two", "three", "four", "five", "six", "seven",
		"eight", "nine", "ten"}

	epRegexps = []regexp.Regexp{
		notInWord(fmt.Sprintf("(?:series|season|s)\\s?(\\d{1,4})(?:\\s(?:.*\\s)?)?(?:episode|ep|e|part|pt)\\s?(\\d{1,3}|%s)(?:\\s?e(\\d{1,2}))*", romanNumeralRe)),
		notInWord(fmt.Sprintf("(?:series|season)\\s?(\\d{1,4})\\s(\\d{1,3})\\s?of\\s?(?:\\d{1,3})")),
		notInWord(fmt.Sprintf("(\\d{1,2})\\s?x\\s?(\\d+)(?:\\s(\\d{1,2}))?")),
		notInWord(fmt.Sprintf("(\\d{1,3})\\s?of\\s?(?:\\d{1,3})")),
		notInWord(fmt.Sprintf("(?:episode|e|ep|part|pt)\\s?(\\d{1,3}|%s)", romanNumeralRe)),
		notInWord(fmt.Sprintf("part\\s(%s)", strings.Join(englishNumbers, "|"))),
	}

	ignorePrefixes = []string{
		"(?:\\[[^\\[.*\\]]*\\])",
		"(?:HD.720p?:)",
		"(?:HD.1080p?:)",
		"(?:HD.2160p?:)",
	}
)

func notInWord(re string) regexp.Regexp {
	return regexp.MustCompile( /*"(?<![^\\W_])"+*/ re /*+"(?![^\\W_])"*/, regexp.CASELESS)
}

//Serie Represent serie object
type Serie struct {
	Name       string
	Episode    int
	Season     int
	Quality    string
	EndEpisode int
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

func (s *SerieParser) guessName(name string) (result Serie, err error) {

	for _, c := range "_.," {
		name = strings.Replace(name, string(c), " ", -1)
	}

	type excludePrefix struct {
		Start string
		Stop  string
	}

	excludePrefixList := []excludePrefix{{Start: "[", Stop: "]"}, {Start: "HD", Stop: "720p"}, {Start: "HD", Stop: "1080p"},}
	for _, e := range excludePrefixList {
		if len(name) > len(e.Start) && name[:len(e.Start)] == e.Start {
			name = strings.TrimSpace(name[strings.Index(name, e.Stop)+len(e.Stop):])
		}

	}

	for _, c := range "[]():" {
		name = strings.Replace(name, string(c), " ", -1)
	}
	matched, matchResult := s.parseIt(name, unwantedRegexps, dummyMatch)
	if matched {
		s.logger.Debugf("Matched %s", matchResult.Matches[0].Value)
		err = errors.New("Matched unwanted names")
		return
	}
	identifiedBy := ""

	matched, matchResult = s.parseIt(name, seasonPackRegexps, s.seasonCB)
	if !matched {
		matched, matchResult = s.parseIt(name, epRegexps, s.episodeCB)
	}
	identifiedBy = "ep"
	if !matched {
		matched, matchResult = s.parseIt(name, dateRegexps, dummyMatch)
		identifiedBy = "date"
	}

	if !matched {
		err = errors.New("No match found")
		return
	}

	extra := ""

	s.logger.Debugf("Found a match %s", matchResult.Matches)
	if matchResult.Matches[0].Index > 1 {
		start := 0
		ignoreReg := regexp.MustCompile(strings.Join(ignorePrefixes, "|"), regexp.CASELESS)
		match := ignoreReg.MatcherString(name, regexp.NOTEMPTY)
		if match.Groups() != 0 {
			start = strings.Index(name, match.GroupString(0))
		}
		extra = name[matchResult.Matches[0].Index:]
		name = name[start : matchResult.Matches[0].Index-1]
		name = strings.Split(name, " - ")[0]
		specialReg := regexp.MustCompile("[\\._\\(\\) ]+", regexp.CASELESS)
		name = string(specialReg.ReplaceAll([]byte(name), []byte(" "), 0))
		name = strings.Trim(name, " -")
		name = strings.ToTitle(name)
	}
	s.logger.Debugf("Identified by %s", identifiedBy)

	result.Quality = ParseQuality(extra, s.logger)
	switch matchResult.context.(type) {
	case *episodeMatch:
		s.logger.Debugf("Matched the episode!!\n")
		em := matchResult.context.(*episodeMatch)
		result.EndEpisode = em.EndEpisode
		result.Episode = em.Episode
		result.Season = em.Season
	}
	result.Name = name

	return
}

type matchCB func(matches matchResult) (bool, interface{})

func dummyMatch(matches matchResult) (bool, interface{}) {
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

func (s *SerieParser) parseIt(name string, regexps []regexp.Regexp, cb matchCB) (bool, matchResult) {
	name = strings.ToLower(name)
	for regID, re := range regexps {

		matches := re.MatcherString(name, regexp.NOTEMPTY)
		if matches.Matches() {
			s.logger.Debugf("Found matches [%d] %s, %v, %v", regID, string(fmt.Sprintf("%s", re)), name, matches)
			nbMatch := 1
			for i := 1; i <= matches.Groups(); i++ {
				if matches.Present(i) {
					nbMatch++
				} else {
					break
				}
			}
			s.logger.Debugf("nbMatch %s", nbMatch)

			res := matchResult{
				Matches: make([]match, nbMatch-1),
			}
			offset := 0
			for i := 1; i < nbMatch; i++ {
				m := matches.GroupString(i)
				mbyte := matches.Group(i)
				s.logger.Debugf("====>%s", mbyte)
				offset += strings.Index(name[offset:], m)
				res.Matches[i-1] = match{
					Value: m,
					Index: offset,
				}
			}
			if matched, context := cb(res); matched {
				res.context = context
				return true, res
			}
		} else {
			s.logger.Debugf("No match for %s %s", re, name)
		}
	}
	return false, matchResult{}
}

func (s *SerieParser) seasonCB(matches matchResult) (bool, interface{}) {
	if len(matches.Matches) == 1 {
		return true, nil
	}
	return false, nil
}

type episodeMatch struct {
	Episode    int
	Season     int
	EndEpisode int
}

func (s *SerieParser) episodeCB(matches matchResult) (bool, interface{}) {
	season := 0
	episode := 0
	endEpisode := 0
	s.logger.Debugf("Found %d matches", len(matches.Matches))
	nbMatches := len(matches.Matches)
	if nbMatches != 0 {
		var epError error
		strEp := ""
		if nbMatches >= 2 {
			strEp = matches.Matches[1].Value
			season, _ = strconv.Atoi(matches.Matches[0].Value)
			episode, epError = strconv.Atoi(strEp)
			if nbMatches == 3 {
				endEpisode, _ = strconv.Atoi(matches.Matches[2].Value)
			}
		} else if nbMatches == 1 {
			season = 1
			strEp = matches.Matches[0].Value
			episode, epError = strconv.Atoi(strEp)
		} else {
			s.logger.Errorf("Unknown matches length %d", nbMatches)
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
			Episode:    episode,
			Season:     season,
			EndEpisode: endEpisode,
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
	return s.guessName(name)
}
