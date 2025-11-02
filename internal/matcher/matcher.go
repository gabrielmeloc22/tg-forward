package matcher

import (
	"regexp"
	"strings"
	"unicode"
)

type Matcher struct {
	patterns []*regexp.Regexp
}

func New(patterns []string) (*Matcher, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return &Matcher{patterns: compiled}, nil
}

func normalizeText(text string) string {
	text = strings.ToLower(text)

	var result strings.Builder
	result.Grow(len(text))

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (m *Matcher) Match(text string) bool {
	normalized := normalizeText(text)
	for _, pattern := range m.patterns {
		if pattern.MatchString(normalized) {
			return true
		}
	}
	return false
}

func (m *Matcher) FindMatches(text string) []string {
	normalized := normalizeText(text)
	var matches []string
	for _, pattern := range m.patterns {
		if pattern.MatchString(normalized) {
			matches = append(matches, pattern.String())
		}
	}
	return matches
}
