package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gabrielmelo/tg-forward/internal/api/router"
	"github.com/gabrielmelo/tg-forward/internal/matcher"
	"github.com/gabrielmelo/tg-forward/internal/repository"
	"github.com/gabrielmelo/tg-forward/internal/service"
	"github.com/go-chi/chi/v5"
)

const TestAPIToken = "test-api-token-12345"

type Fixture struct {
	Router  *chi.Mux
	Service *service.RulesService
	Repo    *repository.RulesRepository
}

func NewFixture(t *testing.T, initialPatterns []string) *Fixture {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.json")

	repo, err := repository.NewRulesRepository(rulesPath)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	if len(initialPatterns) > 0 {
		rules := make([]repository.Rule, len(initialPatterns))
		for i, pattern := range initialPatterns {
			rule, err := repo.AddRule("test-rule", pattern)
			if err != nil {
				t.Fatalf("failed to add initial pattern: %v", err)
			}
			rules[i] = *rule
		}
	}

	m, err := matcher.New(repo.GetPatterns())
	if err != nil {
		t.Fatalf("failed to create matcher: %v", err)
	}

	svc := service.NewRulesService(repo, m)
	r := router.New(svc, TestAPIToken)

	t.Cleanup(func() {
		if err := os.Remove(rulesPath); err != nil && !os.IsNotExist(err) {
			t.Logf("failed to cleanup test file: %v", err)
		}
	})

	return &Fixture{
		Router:  r,
		Service: svc,
		Repo:    repo,
	}
}
