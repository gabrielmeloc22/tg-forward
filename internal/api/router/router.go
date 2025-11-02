package router

import (
	"github.com/gabrielmelo/tg-forward/internal/api/handler"
	"github.com/gabrielmelo/tg-forward/internal/api/middleware"
	"github.com/gabrielmelo/tg-forward/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func New(svc *service.RulesService, apiToken string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)

	healthHandler := handler.NewHealthHandler()
	rulesHandler := handler.NewRulesHandler(svc)

	r.Get("/health", handler.Handler(healthHandler.Handle))

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(apiToken))

		r.Route("/rules", func(r chi.Router) {
			r.Get("/", handler.Handler(rulesHandler.GetRules))
			r.Put("/", handler.HandlerWithBody(rulesHandler.UpdateRules))
			r.Post("/add", handler.HandlerWithBody(rulesHandler.AddRule))
			r.Delete("/remove", handler.HandlerWithBody(rulesHandler.RemoveRule))
		})
	})

	return r
}
