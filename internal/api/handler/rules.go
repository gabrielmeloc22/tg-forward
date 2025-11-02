package handler

import (
	"log"
	"net/http"

	"github.com/gabrielmelo/tg-forward/internal/api/serializer"
	"github.com/gabrielmelo/tg-forward/internal/api/types"
	"github.com/gabrielmelo/tg-forward/internal/service"
)

type RulesHandler struct {
	service *service.RulesService
}

func NewRulesHandler(svc *service.RulesService) *RulesHandler {
	return &RulesHandler{
		service: svc,
	}
}

func (h *RulesHandler) GetRules(w http.ResponseWriter, r *http.Request) (*types.DataResponse, *serializer.Error) {
	rules := h.service.GetRules()
	return &types.DataResponse{Data: types.RulesResponse{Rules: rules}}, nil
}

func (h *RulesHandler) UpdateRules(w http.ResponseWriter, r *http.Request, body *types.UpdateRulesRequest) (*types.DataResponse, *serializer.Error) {
	rules, err := h.service.UpdateRules(body.Rules)
	if err != nil {
		return nil, serializer.NewError(http.StatusBadRequest, "INVALID_RULES", err.Error())
	}

	log.Printf("Rules updated: %d rules", len(rules))
	return &types.DataResponse{Data: types.RulesResponse{Rules: rules}}, nil
}

func (h *RulesHandler) AddRule(w http.ResponseWriter, r *http.Request, body *types.AddRuleRequest) (*types.DataResponse, *serializer.Error) {
	rule, err := h.service.AddRule(body.Name, body.Pattern)
	if err != nil {
		return nil, serializer.NewError(http.StatusBadRequest, "INVALID_RULE", err.Error())
	}

	log.Printf("Rule added: %s (ID: %s)", rule.Name, rule.ID)
	return &types.DataResponse{Data: types.RuleResponse{Rule: *rule}}, nil
}

func (h *RulesHandler) RemoveRule(w http.ResponseWriter, r *http.Request, body *types.RemoveRuleRequest) (*types.DataResponse, *serializer.Error) {
	err := h.service.RemoveRule(body.ID)
	if err != nil {
		return nil, serializer.NewError(http.StatusNotFound, "RULE_NOT_FOUND", err.Error())
	}

	log.Printf("Rule removed: %s", body.ID)
	return &types.DataResponse{Data: map[string]string{"message": "rule deleted successfully"}}, nil
}
