package util

import "strings"

// Tokenize splits a line into words, treating "quoted text" as one
// token even if it contains spaces. Exported (capital T) because
// main.go, which lives in a different package, needs to call it.
func Tokenize(line string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, r := range line {
		switch {
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
