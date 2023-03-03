package parser

import (
	"regexp"
	"strings"
)

var snakeCaseRegex = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase converts strings to snake case in order to have properly
// named parameters, e.g.
// MyOperation -> my-operation
func ToSnakeCase(str string) string {
	snake := snakeCaseRegex.ReplaceAllString(str, "${1}-${2}")
	return strings.ToLower(snake)
}
