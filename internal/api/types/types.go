package types

import "github.com/gabrielmelo/tg-forward/internal/repository"

type DataResponse struct {
	Data interface{} `json:"data"`
}

type RulesResponse struct {
	Rules []repository.Rule `json:"rules"`
}

type RuleResponse struct {
	Rule repository.Rule `json:"rule"`
}

type UpdateRulesRequest struct {
	Rules []repository.Rule `json:"rules"`
}

type AddRuleRequest struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type RemoveRuleRequest struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
