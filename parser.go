package medianame

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
