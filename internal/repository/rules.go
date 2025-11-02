package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type Rule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type Rules struct {
	Rules []Rule `json:"rules"`
}

type RulesRepository struct {
	filePath string
	mu       sync.RWMutex
	rules    *Rules
}

func NewRulesRepository(filePath string) (*RulesRepository, error) {
	repo := &RulesRepository{
		filePath: filePath,
	}

	if err := repo.load(); err != nil {
		if os.IsNotExist(err) {
			repo.rules = &Rules{Rules: []Rule{}}
		} else {
			return nil, fmt.Errorf("failed to load rules: %w", err)
		}
	}

	return repo, nil
}

func (r *RulesRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	var rules Rules
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	r.rules = &rules
	return nil
}

func (r *RulesRepository) save() error {
	data, err := json.MarshalIndent(r.rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	return nil
}

func (r *RulesRepository) GetRules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]Rule, len(r.rules.Rules))
	copy(rules, r.rules.Rules)
	return rules
}

func (r *RulesRepository) SetRules(rules []Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules.Rules = make([]Rule, len(rules))
	copy(r.rules.Rules, rules)

	return r.save()
}

func (r *RulesRepository) AddRule(name, pattern string) (*Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule := Rule{
		ID:      uuid.New().String(),
		Name:    name,
		Pattern: pattern,
	}

	r.rules.Rules = append(r.rules.Rules, rule)
	if err := r.save(); err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *RulesRepository) RemoveRule(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules.Rules {
		if rule.ID == id {
			r.rules.Rules = append(r.rules.Rules[:i], r.rules.Rules[i+1:]...)
			return r.save()
		}
	}

	return fmt.Errorf("rule not found: %s", id)
}

func (r *RulesRepository) GetPatterns() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	patterns := make([]string, len(r.rules.Rules))
	for i, rule := range r.rules.Rules {
		patterns[i] = rule.Pattern
	}
	return patterns
}
