package util

import "strings"

// Tokenize splits a command line into tokens.
//
// Double-quoted text is treated as a single token.
func Tokenize(line string) []string {
	var tokens []string
	var current strings.Builder

	inQuotes := false

	for _, r := range line {
		switch {
		case r == '"':
			inQuotes = !inQuotes

		case (r == ' ' || r == '\t') && !inQuotes:
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
