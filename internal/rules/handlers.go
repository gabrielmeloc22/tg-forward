package rules

import (
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) (*DataResponse, *Error) {
	rules := h.service.GetRules()
	return &DataResponse{Data: RulesResponse{Rules: rules}}, nil
}

func (h *Handler) UpdateRules(w http.ResponseWriter, r *http.Request, body *UpdateRulesRequest) (*DataResponse, *Error) {
	rules, err := h.service.UpdateRules(body.Rules)
	if err != nil {
		return nil, NewError(http.StatusBadRequest, "INVALID_RULES", err.Error())
	}

	log.Printf("Rules updated: %d rules", len(rules))
	return &DataResponse{Data: RulesResponse{Rules: rules}}, nil
}

func (h *Handler) AddRule(w http.ResponseWriter, r *http.Request, body *AddRuleRequest) (*DataResponse, *Error) {
	rule, err := h.service.AddRule(body.Name, body.Pattern, body.Keywords)
	if err != nil {
		return nil, NewError(http.StatusBadRequest, "INVALID_RULE", err.Error())
	}

	log.Printf("Rule added: %s (ID: %s)", rule.Name, rule.ID)
	return &DataResponse{Data: RuleResponse{Rule: *rule}}, nil
}

func (h *Handler) RemoveRule(w http.ResponseWriter, r *http.Request, body *RemoveRuleRequest) (*DataResponse, *Error) {
	err := h.service.RemoveRule(body.ID)
	if err != nil {
		return nil, NewError(http.StatusNotFound, "RULE_NOT_FOUND", err.Error())
	}

	log.Printf("Rule removed: %s", body.ID)
	return &DataResponse{Data: map[string]string{"message": "rule deleted successfully"}}, nil
}
