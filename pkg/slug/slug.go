package slug

import (
	"regexp"
	"strings"
)

var invalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func Make(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = invalidChars.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
