package service

import (
	"fmt"
	"regexp"

	"github.com/gabrielmelo/tg-forward/internal/matcher"
	"github.com/gabrielmelo/tg-forward/internal/repository"
)

type RulesService struct {
	repo    *repository.RulesRepository
	matcher *matcher.Matcher
}

func NewRulesService(repo *repository.RulesRepository, m *matcher.Matcher) *RulesService {
	return &RulesService{
		repo:    repo,
		matcher: m,
	}
}

func (s *RulesService) GetRules() []repository.Rule {
	return s.repo.GetRules()
}

func (s *RulesService) UpdateRules(rules []repository.Rule) ([]repository.Rule, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("at least one rule is required")
	}

	patterns := make([]string, len(rules))
	for i, rule := range rules {
		if err := s.validatePattern(rule.Pattern); err != nil {
			return nil, err
		}
		patterns[i] = rule.Pattern
	}

	if err := s.repo.SetRules(rules); err != nil {
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	newMatcher, err := matcher.New(patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to create matcher: %w", err)
	}
	s.matcher = newMatcher

	return rules, nil
}

func (s *RulesService) AddRule(name, pattern string) (*repository.Rule, error) {
	if err := s.validatePattern(pattern); err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("rule name is required")
	}

	rule, err := s.repo.AddRule(name, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to add rule: %w", err)
	}

	patterns := s.repo.GetPatterns()
	newMatcher, err := matcher.New(patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to update matcher: %w", err)
	}
	s.matcher = newMatcher

	return rule, nil
}

func (s *RulesService) RemoveRule(id string) error {
	if err := s.repo.RemoveRule(id); err != nil {
		return err
	}

	patterns := s.repo.GetPatterns()
	newMatcher, err := matcher.New(patterns)
	if err != nil {
		return fmt.Errorf("failed to update matcher: %w", err)
	}
	s.matcher = newMatcher

	return nil
}

func (s *RulesService) GetMatcher() *matcher.Matcher {
	return s.matcher
}

func (s *RulesService) validatePattern(pattern string) error {
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}
	return nil
}
