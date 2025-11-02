package rules

import (
	"fmt"
	"regexp"

	"github.com/gabrielmelo/tg-forward/internal/matcher"
)

type Service struct {
	repo    *Repository
	matcher *matcher.Matcher
}

func NewService(repo *Repository, m *matcher.Matcher) *Service {
	return &Service{
		repo:    repo,
		matcher: m,
	}
}

func (s *Service) GetRules() []Rule {
	rules, _ := s.repo.GetRules()
	return rules
}

func (s *Service) UpdateRules(rules []Rule) ([]Rule, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("at least one rule is required")
	}

	matchRules := make([]matcher.MatchRule, len(rules))
	for i, rule := range rules {
		if rule.Pattern != "" {
			if err := s.validatePattern(rule.Pattern); err != nil {
				return nil, err
			}
			matchRules[i] = matcher.MatchRule{Pattern: rule.Pattern}
		} else if len(rule.Keywords) > 0 {
			matchRules[i] = matcher.MatchRule{Keywords: rule.Keywords}
		} else {
			return nil, fmt.Errorf("rule must have either pattern or keywords")
		}
	}

	if err := s.repo.SetRules(rules); err != nil {
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	newMatcher, err := matcher.New(matchRules)
	if err != nil {
		return nil, fmt.Errorf("failed to create matcher: %w", err)
	}
	s.matcher = newMatcher

	return rules, nil
}

func (s *Service) AddRule(name, pattern string, keywords []string) (*Rule, error) {
	if pattern != "" {
		if err := s.validatePattern(pattern); err != nil {
			return nil, err
		}
	} else if len(keywords) == 0 {
		return nil, fmt.Errorf("rule must have either pattern or keywords")
	}

	if name == "" {
		return nil, fmt.Errorf("rule name is required")
	}

	rule, err := s.repo.AddRule(name, pattern, keywords)
	if err != nil {
		return nil, fmt.Errorf("failed to add rule: %w", err)
	}

	matchRules, _ := s.repo.GetPatterns()
	newMatcher, err := matcher.New(matchRules)
	if err != nil {
		return nil, fmt.Errorf("failed to update matcher: %w", err)
	}
	s.matcher = newMatcher

	return rule, nil
}

func (s *Service) RemoveRule(id string) error {
	if err := s.repo.RemoveRule(id); err != nil {
		return err
	}

	patterns, _ := s.repo.GetPatterns()
	newMatcher, err := matcher.New(patterns)
	if err != nil {
		return fmt.Errorf("failed to update matcher: %w", err)
	}
	s.matcher = newMatcher

	return nil
}

func (s *Service) GetMatcher() *matcher.Matcher {
	return s.matcher
}

func (s *Service) validatePattern(pattern string) error {
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}
	return nil
}
