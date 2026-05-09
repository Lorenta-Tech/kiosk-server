package routes

import (
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/app"
	"github.com/Lorenta-Tech/kiosk-server/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middlewares.RequestLoggerMiddleware(app.Logger))
	r.Use(middlewares.RecoveryMiddleware(app.Logger))
	r.Use(middlewares.CORSMiddleware)
	r.Use(middleware.RequestSize(5 << 20))

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
			next.ServeHTTP(w, r)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(30 * time.Second))
		r.Get("/health", app.HealthCheck)
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		authRoutes(app, r)
		fileRoutes(app, r)
		paymentRoutes(app, r)
	})

	return r
}

func authRoutes(app *app.Application, r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/google", app.UserHandler.HandleGoogleAuth)
	})
}

func fileRoutes(app *app.Application, r chi.Router) {
	//public routes
	r.Post("/print/jobs/token", app.FileHandler.HandleGetJobByToken)
	r.Route("/files", func(r chi.Router) {
		r.Post("/upload/init", app.FileHandler.HandleInitFileUpload)
		r.Post("/upload/confirm", app.FileHandler.HandleConfirmFileUpload)
		r.Get("/jobs/recent", app.FileHandler.HandleGetRecentPrintJobs)
		r.Get("/jobs/active",app.FileHandler.HandleActivePrintJobs)
	})
}

func paymentRoutes(app *app.Application, r chi.Router) {
	r.Route("/payments", func(r chi.Router) {
		r.Post("/create", app.PaymentHandler.HandleCreateOrder)
	})
	r.Post("/webhooks/razorpay", app.PaymentHandler.HandleWebhook)
}
