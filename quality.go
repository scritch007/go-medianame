package moviename

import (
	"regexp"
)

const (
	resolutionTag = "resolution"
)

var (
	resolution = []qualityComponent{
		qualityComponent{
			t: resolutionTag,
			v: 10,
			n: "360p",
		},
		qualityComponent{t: resolutionTag, v: 20, n: "368p", r: regexp.MustCompile("368p?")},
		qualityComponent{t: resolutionTag, v: 30, n: "480p", r: regexp.MustCompile("480p?")},
		qualityComponent{t: resolutionTag, v: 40, n: "576p", r: regexp.MustCompile("576p?")},
		qualityComponent{t: resolutionTag, v: 45, n: "hr"},
		qualityComponent{t: resolutionTag, v: 50, n: "720i"},
		qualityComponent{t: resolutionTag, v: 60, n: "720p", r: regexp.MustCompile("(1280x)?720(p|hd)?x?(50)?")},
		qualityComponent{t: resolutionTag, v: 70, n: "1080i"},
		qualityComponent{t: resolutionTag, v: 80, n: "1080p", r: regexp.MustCompile("(1920x)?1080p?x?(50)?")},
		qualityComponent{t: resolutionTag, v: 90, n: "2160p", r: regexp.MustCompile("((3840x)?2160p?x?(50)?)|4k")},
	}
)

type qualityComponent struct {
	t string
	v int
	n string
	r *regexp.Regexp
}

func init() {
	for i := range resolution {
		if resolution[i].r == nil {
			resolution[i].r = regexp.MustCompile(resolution[i].n)
		}
	}
}

//ParseQuality return the best quality value
func (m *MovieParser) ParseQuality(name string) string {
	for _, res := range resolution {
		re := res.r
		m.logger.Debug("looking at %s with %s", name, res.n)
		match := re.FindStringSubmatch(name)
		if match != nil {
			return res.n
		}
	}
	return ""
}
