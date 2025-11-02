package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRulesRepository(t *testing.T) {
	t.Run("creates new file if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		rulesPath := filepath.Join(tmpDir, "rules.json")

		repo, err := NewRulesRepository(rulesPath)
		if err != nil {
			t.Fatalf("NewRulesRepository() failed: %v", err)
		}

		if repo == nil {
			t.Fatal("Expected repo to be non-nil")
		}

		if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
			t.Error("Rules file was not created")
		}
	})

	t.Run("loads existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		rulesPath := filepath.Join(tmpDir, "rules.json")

		content := `{"rules": [{"id": "1", "name": "test1", "pattern": "test.*"}, {"id": "2", "name": "test2", "pattern": ".*test"}]}`
		if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		repo, err := NewRulesRepository(rulesPath)
		if err != nil {
			t.Fatalf("NewRulesRepository() failed: %v", err)
		}

		rules := repo.GetRules()
		if len(rules) != 2 {
			t.Errorf("Expected 2 rules, got %d", len(rules))
		}
	})
}

func TestRulesRepository_GetRules(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("NewRulesRepository() failed: %v", err)
	}

	rules := repo.GetRules()
	if rules == nil {
		t.Error("GetRules() returned nil")
	}
}

func TestRulesRepository_GetPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("NewRulesRepository() failed: %v", err)
	}

	repo.AddRule("test1", "pattern1")
	repo.AddRule("test2", "pattern2")

	patterns := repo.GetPatterns()
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}

	if patterns[0] != "pattern1" || patterns[1] != "pattern2" {
		t.Errorf("Patterns mismatch: got %v", patterns)
	}
}

func TestRulesRepository_SetRules(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("NewRulesRepository() failed: %v", err)
	}

	newRules := []Rule{
		{ID: "1", Name: "rule1", Pattern: "pattern1"},
		{ID: "2", Name: "rule2", Pattern: "pattern2"},
		{ID: "3", Name: "rule3", Pattern: "pattern3"},
	}
	if err := repo.SetRules(newRules); err != nil {
		t.Errorf("SetRules() failed: %v", err)
	}

	rules := repo.GetRules()
	if len(rules) != len(newRules) {
		t.Errorf("Expected %d rules, got %d", len(newRules), len(rules))
	}

	for i, r := range rules {
		if r.ID != newRules[i].ID || r.Name != newRules[i].Name || r.Pattern != newRules[i].Pattern {
			t.Errorf("Rule %d: got %+v, want %+v", i, r, newRules[i])
		}
	}
}

func TestRulesRepository_AddRule(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("NewRulesRepository() failed: %v", err)
	}

	initialCount := len(repo.GetRules())

	rule, err := repo.AddRule("new_rule", "new_pattern")
	if err != nil {
		t.Errorf("AddRule() failed: %v", err)
	}

	if rule.Name != "new_rule" || rule.Pattern != "new_pattern" {
		t.Errorf("AddRule() returned wrong rule: %+v", rule)
	}

	if rule.ID == "" {
		t.Error("AddRule() did not generate ID")
	}

	rules := repo.GetRules()
	if len(rules) != initialCount+1 {
		t.Errorf("Expected %d rules, got %d", initialCount+1, len(rules))
	}

	if rules[len(rules)-1].Pattern != "new_pattern" {
		t.Errorf("Last rule pattern = %s, want new_pattern", rules[len(rules)-1].Pattern)
	}
}

func TestRulesRepository_RemoveRule(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("NewRulesRepository() failed: %v", err)
	}

	rules := []Rule{
		{ID: "1", Name: "rule1", Pattern: "pattern1"},
		{ID: "2", Name: "rule2", Pattern: "pattern2"},
		{ID: "3", Name: "rule3", Pattern: "pattern3"},
	}
	repo.SetRules(rules)

	if err := repo.RemoveRule("2"); err != nil {
		t.Errorf("RemoveRule() failed: %v", err)
	}

	remainingRules := repo.GetRules()
	if len(remainingRules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(remainingRules))
	}

	for _, r := range remainingRules {
		if r.ID == "2" {
			t.Error("rule with ID 2 was not removed")
		}
	}

	if err := repo.RemoveRule("nonexistent"); err == nil {
		t.Error("Expected error when removing nonexistent rule")
	}
}
