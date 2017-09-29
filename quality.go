package medianame

import (
	"regexp"
	"strings"

	"github.com/labstack/gommon/log"
)

const (
	resolutionTag = "resolution"
	sourceTag     = "source"
	codecTag      = "codec"
)

var (
	resolutions = []qualityComponent{
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

	sources = []qualityComponent{
		qualityComponent{t: sourceTag, v: 10, n: "workprint", m: -8},
		qualityComponent{t: sourceTag, v: 20, n: "cam", r: regexp.MustCompile("(?:hd)?cam"), m: -7},
		qualityComponent{t: sourceTag, v: 30, n: "ts", r: regexp.MustCompile("(?:hd)?ts|telesync"), m: -6},
		qualityComponent{t: sourceTag, v: 40, n: "tc", r: regexp.MustCompile("tc|telecine"), m: -5},
		qualityComponent{t: sourceTag, v: 50, n: "r5", r: regexp.MustCompile("r[2-8c]"), m: -4},
		qualityComponent{t: sourceTag, v: 60, n: "hdrip", r: regexp.MustCompile("hd[\\W_]?rip"), m: -3},
		qualityComponent{t: sourceTag, v: 70, n: "ppvrip", r: regexp.MustCompile("ppv[\\W_]?rip"), m: -2},
		qualityComponent{t: sourceTag, v: 80, n: "preair", m: -1},
		qualityComponent{t: sourceTag, v: 90, n: "tvrip", r: regexp.MustCompile("tv[\\W_]?rip")},
		qualityComponent{t: sourceTag, v: 100, n: "dsr", r: regexp.MustCompile("dsr|ds[\\W_]?rip")},
		qualityComponent{t: sourceTag, v: 110, n: "sdtv", r: regexp.MustCompile("(?:[sp]dtv|dvb)(?:[\\W_]?rip)?")},
		qualityComponent{t: sourceTag, v: 120, n: "dvdscr", r: regexp.MustCompile("(?:(?:dvd|web)[\\W_]?)?scr(?:eener)?"), m: 0},
		qualityComponent{t: sourceTag, v: 130, n: "bdscr", r: regexp.MustCompile("bdscr(?:eener)?")},
		qualityComponent{t: sourceTag, v: 140, n: "webrip", r: regexp.MustCompile("web[\\W_]?rip")},
		qualityComponent{t: sourceTag, v: 150, n: "hdtv", r: regexp.MustCompile("a?hdtv(?:[\\W_]?rip)?")},
		qualityComponent{t: sourceTag, v: 160, n: "webdl", r: regexp.MustCompile("web(?:[\\W_]?(dl|hd))?")},
		qualityComponent{t: sourceTag, v: 170, n: "dvdrip", r: regexp.MustCompile("dvd(?:[\\W_]?rip)?")},
		qualityComponent{t: sourceTag, v: 175, n: "remux"},
		qualityComponent{t: sourceTag, v: 180, n: "bluray", r: regexp.MustCompile("(?:b[dr][\\W_]?rip|blu[\\W_]?ray(?:[\\W_]?rip)?)")},
	}
	codecs = []qualityComponent{
		qualityComponent{t: codecTag, v: 10, n: "divx"},
		qualityComponent{t: codecTag, v: 20, n: "xvid"},
		qualityComponent{t: codecTag, v: 30, n: "h264", r: regexp.MustCompile("[hx].?264")},
		qualityComponent{t: codecTag, v: 35, n: "vp9"},
		qualityComponent{t: codecTag, v: 40, n: "h265", r: regexp.MustCompile("[hx].?265|hevc")},
		qualityComponent{t: codecTag, v: 50, n: "10bit", r: regexp.MustCompile("10.?bit|hi10p")},
	}
)

type qualityComponent struct {
	t string         //Type
	v int            //Value
	n string         //Name
	r *regexp.Regexp //Regexp
	m int            //Modifier
}

func init() {
	for i := range resolutions {
		if resolutions[i].r == nil {
			resolutions[i].r = regexp.MustCompile(resolutions[i].n)
		}
	}

	for i := range sources {
		if sources[i].r == nil {
			sources[i].r = regexp.MustCompile(sources[i].n)
		}
	}

	for i := range codecs {
		if codecs[i].r == nil {
			codecs[i].r = regexp.MustCompile(codecs[i].n)
		}
	}
}

//ParseQuality return the best quality value
func ParseQuality(name string, logger *log.Logger) string {
	name = strings.ToLower(name)
	resolution := ""
	source := ""
	for _, res := range resolutions {
		re := res.r
		logger.Debugf("looking at %s with %s", name, res.n)
		match := re.FindStringSubmatch(name)
		if match != nil {
			resolution = res.n
			break
		}
	}
	for _, res := range sources {
		re := res.r
		logger.Debugf("looking at %s with %s", name, res.n)
		match := re.FindStringSubmatch(name)
		if match != nil {
			source = res.n
			break
		}
	}
	return strings.Join([]string{resolution, source}, " ")
}
