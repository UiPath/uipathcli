package parser

import (
	"regexp"
	"strings"
)

var snakeCaseRegex = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := snakeCaseRegex.ReplaceAllString(str, "${1}-${2}")
	return strings.ToLower(snake)
}
