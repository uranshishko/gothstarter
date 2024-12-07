package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/uranshishko/gothstarter/auth"
	"github.com/uranshishko/gothstarter/common"
)

func NewAuthHandler(r *chi.Mux) {
	r.Get("/login", common.Make(func(hc common.HandlerContext) error {
		if _, err := auth.Client.CompleteAuth(hc); err == nil {
			hc.Redirect(302, "/")
			return nil
		}

		auth.Client.BeginAuth(hc)
		return nil
	}))

	r.Get("/callback", common.Make(func(hc common.HandlerContext) error {
		_, err := auth.Client.CompleteAuth(hc)
		if err != nil {
			hc.Redirect(302, "/login")
			return nil
		}

		hc.Redirect(302, "/")
		return nil
	}))

	r.Get("/logout", common.Make(func(hc common.HandlerContext) error {
		err := auth.Client.Logout(hc)
		if err != nil {
			return err
		}

		hc.Redirect(302, "/login")
		return nil
	}))
}
