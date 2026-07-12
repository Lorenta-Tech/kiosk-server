package routes

import (
	"net/http"
	"time"
	//"fmt"
	"github.com/Lorenta-Tech/kiosk-server/internal/app"
	"github.com/Lorenta-Tech/kiosk-server/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	appjwt "github.com/Lorenta-Tech/kiosk-server/pkg/jwt"
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
			w.Header().Set(
				"X-Request-ID",
				middleware.GetReqID(r.Context()),
			)
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
		notesRoutes(app, r)

		// IMPORTANT
		deptadminRoutes(app, r)
        adminRoutes(app,r)
	})

	return r
}

func authRoutes(app *app.Application, r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/google", app.UserHandler.HandleGoogleAuth)
	})
}

func fileRoutes(app *app.Application, r chi.Router) {

	r.Post("/print/jobs/token", app.FileHandler.HandleGetJobByToken)
	r.Post("/print/jobs/error", app.FileHandler.HandleErrorRequestFromPrinter)
	r.Post("/print/jobs/expire", app.FileHandler.HandleExpireSessionAfterPrinting)
	//r.Get("/admin/getprintjobs",app.FileHandler.HandleFetchPrintJobs)
	r.Route("/files", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(app.JWTSecret,app.Logger))
		r.Post("/upload/init", app.FileHandler.HandleInitFileUpload)
		r.Post("/upload/confirm", app.FileHandler.HandleConfirmFileUpload)
		r.Get("/jobs/recent", app.FileHandler.HandleGetRecentPrintJobs)
		r.Get("/jobs/active", app.FileHandler.HandleActivePrintJobs)
		r.Get("/job/session/:session_id", app.FileHandler.HandleGetJobBySessionID)
	})
}

func paymentRoutes(app *app.Application, r chi.Router) {
	r.Route("/payments", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(app.JWTSecret,app.Logger))
		r.Post("/create", app.PaymentHandler.HandleCreateOrder)
		r.Get("/status/{session_id}", app.PaymentHandler.HandleGetPaymentStatus)
	})
	r.Post("/webhooks/razorpay", app.PaymentHandler.HandleWebhook)
}

// func notesRoutes(app *app.Application, r chi.Router) {

//     // Admin routes — NO auth middleware during testing
//     // TODO: add dept admin auth middleware before production
//     r.Route("/admin", func(r chi.Router) {
//         r.Post("/notes/upload/init",    app.NotesHandler.HandleInitNoteUpload)
//         r.Post("/notes/upload/confirm", app.NotesHandler.HandleConfirmNoteUpload)
//         r.Put("/notes/{note_id}",       app.NotesHandler.HandleUpdateNote)
//         r.Delete("/notes/{note_id}",    app.NotesHandler.HandleDeleteNote)//todo: hard delete
//         r.Post("/subjects",             app.NotesHandler.HandleCreateSubject)
//     })

//     // Student routes — keep auth middleware, students use Google JWT
//     r.Route("/notes", func(r chi.Router) {
//         r.Use(middlewares.AuthMiddleware(app.JWTSecret))
//         r.Get("/branches",                           app.NotesHandler.HandleListBranches)
//         r.Get("/branches/{branch_id}/semesters",     app.NotesHandler.HandleListSemesters)
//         r.Get("/semesters/{semester_id}/subjects",   app.NotesHandler.HandleListSubjects)
//         r.Get("/subjects/{subject_id}/modules",      app.NotesHandler.HandleListModules)
//         r.Get("/modules/{module_id}/notes",          app.NotesHandler.HandleListNotes)
//         r.Get("/{note_id}/view",                     app.NotesHandler.HandleGetNoteViewURL)
//         r.Post("/print/init",                        app.NotesHandler.HandleNotesPrintInit)
//     })
// }

func notesRoutes(app *app.Application, r chi.Router) {

	r.Route("/notes", func(r chi.Router) {

		r.Use(middlewares.AuthMiddleware(app.JWTSecret, app.Logger))

		r.Get("/branches", app.NotesHandler.HandleListBranches)

		r.Get(
			"/branches/{branch_id}/semesters",
			app.NotesHandler.HandleListSemesters,
		)

		r.Get(
			"/semesters/{semester_id}/subjects",
			app.NotesHandler.HandleListSubjects,
		)

		r.Get(
			"/subjects/{subject_id}/modules",
			app.NotesHandler.HandleListModules,
		)

		r.Get(
			"/modules/{module_id}/notes",
			app.NotesHandler.HandleListNotes,
		)

		r.Get(
			"/{note_id}/view",
			app.NotesHandler.HandleGetNoteViewURL,
		)

		r.Post(
			"/print/init",
			app.NotesHandler.HandleNotesPrintInit,
		)
	})
}

func deptadminRoutes(app *app.Application, r chi.Router) {

	secrets := middlewares.Secrets{
		DeptAdmin:  app.JWTSecret,
		SuperAdmin: app.JWTSecret,
	}

	r.Route("/deptadmin", func(r chi.Router) {  //getting changed need to inform frontend team

		r.Route("/auth", func(r chi.Router) {
			r.Post("/super/login", app.DeptAdminHandler.HandleSuperAdminLogin)
			r.Post("/login", app.DeptAdminHandler.HandleDeptAdminLogin)
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole(secrets, app.Logger, appjwt.RoleSuperAdmin))
			r.Post("/dept-admins", app.DeptAdminHandler.HandleRegisterDeptAdmin)
			r.Get("/dept-admins", app.DeptAdminHandler.HandleListDeptAdmins)
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole(secrets, app.Logger, appjwt.RoleDeptAdmin, appjwt.RoleSuperAdmin))

			// Write routes — notes/subjects management
			r.Post("/notes/upload/init", app.NotesHandler.HandleInitNoteUpload)
			r.Post("/notes/upload/confirm", app.NotesHandler.HandleConfirmNoteUpload)
			r.Put("/notes/{note_id}", app.NotesHandler.HandleUpdateNote)
			r.Delete("/notes/{note_id}", app.NotesHandler.HandleDeleteNote)
			r.Post("/subjects", app.NotesHandler.HandleCreateSubject)

			// NEW — read-only catalog browsing, admin-side.
			// Same handlers as the student /notes/* routes, mounted again
			// here so dept admins (and super admin) can browse the same
			// catalog without needing a student Google-OAuth token.
			r.Get("/branches", app.NotesHandler.HandleListBranches)
			r.Get("/branches/{branch_id}/semesters", app.NotesHandler.HandleListSemesters)
			r.Get("/semesters/{semester_id}/subjects", app.NotesHandler.HandleListSubjects)
			r.Get("/subjects/{subject_id}/modules", app.NotesHandler.HandleListModules)
			r.Get("/modules/{module_id}/notes", app.NotesHandler.HandleListNotes)
		})
	})
}
    
	
func adminRoutes(app *app.Application, r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Get("/print/history", app.AdminHandler.HandleFetchPrintHistory)
		r.Get("/print/revenue", app.AdminHandler.HandleGetTotalRevenue)
		r.Get("/print/totalsheetsprinted",app.AdminHandler.HandleGetTotalSheetsPrinted)
		r.Get("/print/colorsheets",app.AdminHandler.HandleGetTotalColorSheetsPrinted)
		r.Get("/print/blackandwhite",app.AdminHandler.HandleGetTotalBlackAndWhiteSheetsPrinted)
		r.Get("/print/revenue-24h",app.AdminHandler.HandleGetRevenueLast24Hours)
		r.Get("/print/sheets-24h",app.AdminHandler.HandleGetSheetsPrintedLast24Hours)
		r.Get("/print/color-sheets-24h",app.AdminHandler.HandleGetColorSheetsPrintedLast24Hours)
		r.Get("/print/black-and-white-sheets-24h",app.AdminHandler.HandleGetBlackAndWhiteSheetsPrintedLast24Hours)
	})
}