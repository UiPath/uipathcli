package parser

import (
	"strings"
	"unicode"
)

// toSnakeCase converts strings to snake case in order to have properly
// named parameters, e.g.
// MyOperation -> my-operation
func toSnakeCase(name string) string {
	var builder strings.Builder
	previousChar := ' '
	withSeparator := false
	for _, char := range name {
		if char == '{' || char == '}' || char == '$' {
			continue
		}
		if char == '/' || char == '_' || char == '-' {
			withSeparator = true
			continue
		}
		if unicode.IsLower(previousChar) && unicode.IsUpper(char) {
			withSeparator = true
		}

		if withSeparator {
			builder.WriteRune('-')
		}
		builder.WriteRune(unicode.ToLower(char))

		withSeparator = false
		previousChar = char
	}
	return builder.String()
}
