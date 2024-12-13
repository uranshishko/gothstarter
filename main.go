package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/uranshishko/gothstarter/auth"
	"github.com/uranshishko/gothstarter/common"
	"github.com/uranshishko/gothstarter/handlers"
	"github.com/uranshishko/gothstarter/middleware"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	mode := os.Getenv("GO_ENV")

	var (
		azureadClient      = os.Getenv("AZUREAD_CLIENT")
		azureadSecret      = os.Getenv("AZUREAD_SECRET")
		azureadCallbackUrl = os.Getenv("AZUREAD_CALLBACK_URL")
		azureadTenant      = os.Getenv("AZUREAD_TENANT")

		sessionSecret = os.Getenv("SESSION_SECRET")
	)

	auth.NewMsalClient(
		azureadTenant,
		azureadClient,
		azureadSecret,
		azureadCallbackUrl,
	)

	// set custom session name
	// auth.SessionName = "session_name"

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = common.IIf(mode == "production", true, false)

	auth.Client.Store = store
}

func main() {
	router := chi.NewMux()

	// Public assets
	router.Handle("/*", public())

	// Public Pages
	router.Handle("/login", common.Make(handlers.LoginHandler))

	// Protected pages
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/", common.Make(handlers.HomeHandler))
	})

	// Auth
	auth := chi.NewRouter()
	handlers.NewAuthHandler(auth)
	router.Mount("/auth", auth)

	// API
	v1 := chi.NewRouter()
	router.Mount("/api/v1", v1)

	listenAddr := flag.String("listenAddr", ":8080", "HTTP listen address")
	slog.Info("HTTP server started", "listenAddr", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, router); err != nil {
		log.Fatal(err)
	}
}
