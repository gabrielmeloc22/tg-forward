package router_test

import (
	"net/http"
	"testing"

	"github.com/gabrielmelo/tg-forward/internal/api/serializer"
	"github.com/gabrielmelo/tg-forward/internal/api/types"
	"github.com/gabrielmelo/tg-forward/internal/repository"
	"github.com/gabrielmelo/tg-forward/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	fixture := testutils.NewFixture(t, nil)

	t.Run("should return ok status without authentication", func(t *testing.T) {
		req := testutils.NewRequest(t, "GET", "/health", nil)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)
		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "ok", dataMap["status"])
	})
}

func TestAuthenticationMiddleware(t *testing.T) {
	t.Parallel()

	fixture := testutils.NewFixture(t, nil)

	t.Run("should return 401 when no token provided", func(t *testing.T) {
		req := testutils.NewRequest(t, "GET", "/rules", nil)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Equal(t, "UNAUTHORIZED", body.Code)
	})

	t.Run("should return 401 when invalid token provided", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, "wrong-token")

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		require.Equal(t, "UNAUTHORIZED", body.Code)
	})

	t.Run("should return 200 when valid token provided", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testutils.TestAPIToken)

		res := testutils.ExecuteRequest(req, fixture.Router)

		require.Equal(t, http.StatusOK, res.Code)
	})
}

func TestGetRulesHandler(t *testing.T) {
	t.Parallel()

	initialPatterns := []string{"test.*", "urgent", "important"}
	fixture := testutils.NewFixture(t, initialPatterns)

	t.Run("should return all rules", func(t *testing.T) {
		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testutils.TestAPIToken)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, len(initialPatterns), len(rulesArray))
	})

	t.Run("should return empty array when no rules exist", func(t *testing.T) {
		emptyFixture := testutils.NewFixture(t, nil)

		req := testutils.NewAuthenticatedRequest(t, "GET", "/rules", nil, testutils.TestAPIToken)

		res := testutils.ExecuteRequest(req, emptyFixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, 0, len(rulesArray))
	})
}

func TestAddRuleHandler(t *testing.T) {
	t.Parallel()

	initialPatterns := []string{"test.*"}
	fixture := testutils.NewFixture(t, initialPatterns)

	t.Run("should add new rule", func(t *testing.T) {
		reqBody := types.AddRuleRequest{Name: "Important Messages", Pattern: "important.*"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

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
		reqBody := types.AddRuleRequest{Name: "Bad Rule", Pattern: "[unclosed"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULE", body.Code)
	})

	t.Run("should return 400 when name is missing", func(t *testing.T) {
		reqBody := types.AddRuleRequest{Pattern: "test.*"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"POST",
			"/rules/add",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULE", body.Code)
	})
}

func TestRemoveRuleHandler(t *testing.T) {
	t.Parallel()

	initialPatterns := []string{"test.*", "urgent"}
	fixture := testutils.NewFixture(t, initialPatterns)

	t.Run("should remove existing rule", func(t *testing.T) {
		rules := fixture.Repo.GetRules()
		require.NotEmpty(t, rules)

		ruleToRemove := rules[0]

		reqBody := types.RemoveRuleRequest{ID: ruleToRemove.ID}

		req := testutils.NewAuthenticatedRequest(
			t,
			"DELETE",
			"/rules/remove",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "rule deleted successfully", dataMap["message"])

		remainingRules := fixture.Repo.GetRules()
		require.Equal(t, len(rules)-1, len(remainingRules))
	})

	t.Run("should return 404 for nonexistent rule", func(t *testing.T) {
		reqBody := types.RemoveRuleRequest{ID: "nonexistent-id"}

		req := testutils.NewAuthenticatedRequest(
			t,
			"DELETE",
			"/rules/remove",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusNotFound, res.Code)
		require.Equal(t, "RULE_NOT_FOUND", body.Code)
	})
}

func TestUpdateRulesHandler(t *testing.T) {
	t.Parallel()

	fixture := testutils.NewFixture(t, []string{"old.*"})

	t.Run("should update rules with valid rules", func(t *testing.T) {
		newRules := []repository.Rule{
			{ID: "1", Name: "New Rule 1", Pattern: "new.*"},
			{ID: "2", Name: "New Rule 2", Pattern: "pattern[0-9]+"},
		}
		reqBody := types.UpdateRulesRequest{Rules: newRules}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[types.DataResponse](t, res.Body)

		require.Equal(t, http.StatusOK, res.Code)

		dataMap, ok := body.Data.(map[string]interface{})
		require.True(t, ok)

		rulesArray, ok := dataMap["rules"].([]interface{})
		require.True(t, ok)
		require.Equal(t, len(newRules), len(rulesArray))
	})

	t.Run("should return 400 for invalid regex pattern", func(t *testing.T) {
		reqBody := types.UpdateRulesRequest{Rules: []repository.Rule{
			{ID: "1", Name: "Bad Rule", Pattern: "[invalid("},
		}}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULES", body.Code)
	})

	t.Run("should return 400 for empty rules array", func(t *testing.T) {
		reqBody := types.UpdateRulesRequest{Rules: []repository.Rule{}}

		req := testutils.NewAuthenticatedRequest(
			t,
			"PUT",
			"/rules",
			testutils.MarshallBody(t, reqBody),
			testutils.TestAPIToken,
		)

		res := testutils.ExecuteRequest(req, fixture.Router)

		body := testutils.UnmarshallReqBody[serializer.ApiErrorResponse](t, res.Body)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Equal(t, "INVALID_RULES", body.Code)
	})
}
