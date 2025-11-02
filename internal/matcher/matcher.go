package matcher

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type MatchRule struct {
	Pattern  string
	Keywords []string
}

type Matcher struct {
	patterns       []*regexp.Regexp
	keywordMatches [][]string
}

func New(rules []MatchRule) (*Matcher, error) {
	patterns := make([]*regexp.Regexp, 0)
	keywords := make([][]string, 0)

	for _, rule := range rules {
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, re)
			keywords = append(keywords, nil)
		} else if len(rule.Keywords) > 0 {
			patterns = append(patterns, nil)
			keywords = append(keywords, rule.Keywords)
		}
	}

	return &Matcher{
		patterns:       patterns,
		keywordMatches: keywords,
	}, nil
}

func normalizeText(text string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	text, _, _ = transform.String(t, text)

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

	for i := range m.patterns {
		if m.patterns[i] != nil {
			if m.patterns[i].MatchString(normalized) {
				return true
			}
		} else if m.keywordMatches[i] != nil {
			if matchesAllKeywords(normalized, m.keywordMatches[i]) {
				return true
			}
		}
	}
	return false
}

func (m *Matcher) FindMatches(text string) []string {
	normalized := normalizeText(text)
	var matches []string

	for i := range m.patterns {
		if m.patterns[i] != nil {
			if m.patterns[i].MatchString(normalized) {
				matches = append(matches, m.patterns[i].String())
			}
		} else if m.keywordMatches[i] != nil {
			if matchesAllKeywords(normalized, m.keywordMatches[i]) {
				matches = append(matches, strings.Join(m.keywordMatches[i], ", "))
			}
		}
	}
	return matches
}

func matchesAllKeywords(text string, keywords []string) bool {
	for _, keyword := range keywords {
		normalizedKeyword := normalizeText(keyword)
		if !strings.Contains(text, normalizedKeyword) {
			return false
		}
	}
	return true
}
