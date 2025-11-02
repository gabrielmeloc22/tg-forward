package matcher

import "regexp"

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

func (m *Matcher) Match(text string) bool {
	for _, pattern := range m.patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (m *Matcher) FindMatches(text string) []string {
	var matches []string
	for _, pattern := range m.patterns {
		if pattern.MatchString(text) {
			matches = append(matches, pattern.String())
		}
	}
	return matches
}
