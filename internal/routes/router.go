package routes

import (
	"net/http"

	"github.com/Lorenta-Tech/kiosk-server/internal/app"
	"github.com/Lorenta-Tech/kiosk-server/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	// Global Middlwares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middlewares.RequestLoggerMiddleware(app.Logger))
	r.Use(middlewares.RecoveryMiddleware(app.Logger))
	r.Use(middlewares.CORSMiddleware)
	r.Use(middleware.RequestSize(5 << 20))

	//response headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
			next.ServeHTTP(w, r)
		})
	})

	r.Group(func(r chi.Router) {
		r.Get("/health", app.HealthCheck)
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})

	return r
}
