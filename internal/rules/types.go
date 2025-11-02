package rules

type DataResponse struct {
	Data any `json:"data"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type RulesResponse struct {
	Rules []Rule `json:"rules"`
}

type RuleResponse struct {
	Rule Rule `json:"rule"`
}

type UpdateRulesRequest struct {
	Rules []Rule `json:"rules"`
}

type AddRuleRequest struct {
	Name     string   `json:"name"`
	Pattern  string   `json:"pattern"`
	Keywords []string `json:"keywords"`
}

type RemoveRuleRequest struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
