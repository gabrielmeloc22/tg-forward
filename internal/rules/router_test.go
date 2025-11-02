package rules_test

import (
	"net/http"
	"testing"

	"github.com/gabrielmelo/tg-forward/internal/rules"
	"github.com/gabrielmelo/tg-forward/internal/testutils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

const testAPIToken = "test-api-token-12345"

func setupRouter(t *testing.T, initialPatterns []string) (*chi.Mux, *rules.Repository, func()) {
	client, database, cleanup := testutils.SetupTestDB(t)
	fixture := testutils.NewFixture(t, client, database, testAPIToken, initialPatterns)

	return fixture.Router, fixture.RulesRepo, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	r, _, cleanup := setupRouter(t, nil)
	defer cleanup()

	t.Run("should return ok status without authentication", func(t *testing.T) {
		req := testutils.NewRequest(t, "GET", "/health", nil)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)
		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "ok", dataMap["status"])
	})
}

func TestAuthenticationMiddleware(t *testing.T) {
	r, _, cleanup := setupRouter(t, nil)
	defer cleanup()

	t.Run("should return 401 when no token provided", func(t *testing.T) {
		req := testutils.NewRequest(t, "GET", "/rules", nil)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Equal(t, "UNAUTHORIZED", body.Code)
	})

	t.Run("should return 401 when invalid token provided", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, "wrong-token")

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Equal(t, "UNAUTHORIZED", body.Code)
	})

	t.Run("should return 200 when valid token provided", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testAPIToken)

		res := testutils.ExecuteRequest(req, r)

		require.Equal(t, http.StatusOK, res.Code)
	})
}

func TestGetRulesHandler(t *testing.T) {
	initialPatterns := []string{"test.*", "urgent", "important"}
	r, _, cleanup := setupRouter(t, initialPatterns)
	defer cleanup()

	t.Run("should return all rules", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testAPIToken)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, len(initialPatterns), len(rulesArray))
	})

	t.Run("should return empty array when no rules exist", func(t *testing.T) {
		emptyR, _, emptyCleanup := setupRouter(t, nil)
		defer emptyCleanup()

		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testAPIToken)

		res := testutils.ExecuteRequest(req, emptyR)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, 0, len(rulesArray))
	})
}

func TestAddRuleHandler(t *testing.T) {
	initialPatterns := []string{"test.*"}
	r, _, cleanup := setupRouter(t, initialPatterns)
	defer cleanup()

	t.Run("should add new rule", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{Name: "Important Messages", Pattern: "important.*"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		ruleMap, ok := dataMap["rule"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Important Messages", ruleMap["name"])
		require.Equal(t, "important.*", ruleMap["pattern"])
		require.NotEmpty(t, ruleMap["id"])
	})

	t.Run("should return 400 for invalid regex pattern", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{Name: "Bad Rule", Pattern: "[unclosed"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULE", body.Code)
	})

	t.Run("should return 400 when name is missing", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{Pattern: "test.*"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULE", body.Code)
	})

	t.Run("should add rule with keywords only", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{
			Name:     "Keyword Rule",
			Keywords: []string{"urgent", "alert"},
		}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		ruleMap, ok := dataMap["rule"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Keyword Rule", ruleMap["name"])
		require.NotEmpty(t, ruleMap["id"])

		keywordsArray, ok := ruleMap["keywords"].([]interface{})
		require.True(t, ok)
		require.Equal(t, 2, len(keywordsArray))
	})

	t.Run("should add rule with both pattern and keywords", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{
			Name:     "Mixed Rule",
			Pattern:  "important.*",
			Keywords: []string{"critical", "urgent"},
		}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		ruleMap, ok := dataMap["rule"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Mixed Rule", ruleMap["name"])
		require.Equal(t, "important.*", ruleMap["pattern"])
		require.NotEmpty(t, ruleMap["id"])

		keywordsArray, ok := ruleMap["keywords"].([]interface{})
		require.True(t, ok)
		require.Equal(t, 2, len(keywordsArray))
	})

	t.Run("should return 400 when neither pattern nor keywords provided", func(t *testing.T) {
		reqBody := rules.AddRuleRequest{Name: "Empty Rule"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULE", body.Code)
	})
}

func TestRemoveRuleHandler(t *testing.T) {
	initialPatterns := []string{"test.*", "urgent"}
	r, repo, cleanup := setupRouter(t, initialPatterns)
	defer cleanup()

	t.Run("should remove existing rule", func(t *testing.T) {
		existingRules, err := repo.GetRules()
		require.NoError(t, err)
		require.NotEmpty(t, existingRules)

		ruleToRemove := existingRules[0]

		reqBody := rules.RemoveRuleRequest{ID: ruleToRemove.ID}

		req := testutils.NewAuthenticatedRequest(
			t,
			"DELETE",
			"/rules/remove",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "rule deleted successfully", dataMap["message"])

		remainingRules, err := repo.GetRules()
		require.NoError(t, err)
		require.Equal(t, len(existingRules)-1, len(remainingRules))
	})

	t.Run("should return 404 for nonexistent rule", func(t *testing.T) {
		reqBody := rules.RemoveRuleRequest{ID: "nonexistent-id"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"DELETE",
			"/rules/remove",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusNotFound, res.Code)
		require.Equal(t, "RULE_NOT_FOUND", body.Code)
	})
}

func TestUpdateRulesHandler(t *testing.T) {
	r, _, cleanup := setupRouter(t, []string{"old.*"})
	defer cleanup()

	t.Run("should update rules with valid rules", func(t *testing.T) {
		newRules := []rules.Rule{
			{ID: "1", Name: "New Rule 1", Pattern: "new.*"},
			{ID: "2", Name: "New Rule 2", Pattern: "pattern[0-9]+"},
		}
		reqBody := rules.UpdateRulesRequest{Rules: newRules}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, len(newRules), len(rulesArray))
	})

	t.Run("should return 400 for invalid regex pattern", func(t *testing.T) {
		reqBody := rules.UpdateRulesRequest{Rules: []rules.Rule{
			{ID: "1", Name: "Bad Rule", Pattern: "[invalid("},
		}}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULES", body.Code)
	})

	t.Run("should return 400 for empty rules array", func(t *testing.T) {
		reqBody := rules.UpdateRulesRequest{Rules: []rules.Rule{}}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testAPIToken,
		)

		res := testutils.ExecuteRequest(req, r)

		body := testutils.UnmarshallReqBody[rules.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULES", body.Code)
	})
}
