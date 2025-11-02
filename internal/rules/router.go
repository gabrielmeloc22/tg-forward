package rules

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabrielmelo/tg-forward/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func healthHandler(w http.ResponseWriter, r *http.Request) (*DataResponse, *Error) {
	return &DataResponse{Data: HealthResponse{Status: "ok"}}, nil
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	htmlPath := filepath.Join(wd, "web", "admin.html")
	http.ServeFile(w, r, htmlPath)
}

func NewRouter(svc *Service, apiToken string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)

	rulesHandler := NewHandler(svc)

	r.Get("/health", Wrap(healthHandler))
	r.Get("/admin", adminHandler)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(apiToken))

		r.Route("/rules", func(r chi.Router) {
			r.Get("/", Wrap(rulesHandler.GetRules))
			r.Put("/", WrapWithBody(rulesHandler.UpdateRules))
			r.Post("/add", WrapWithBody(rulesHandler.AddRule))
			r.Delete("/remove", WrapWithBody(rulesHandler.RemoveRule))
			r.Patch("/{id}", WrapWithBodyAndID(rulesHandler.UpdateRule))
		})
	})

	return r
}
