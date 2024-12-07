package handlers

import (
	"github.com/uranshishko/gothstarter/auth"
	"github.com/uranshishko/gothstarter/common"
	"github.com/uranshishko/gothstarter/views/pages"
)

func LoginHandler(hc common.HandlerContext) error {
	if _, err := auth.Client.CompleteAuth(hc); err == nil {
		hc.Redirect(302, "/")
		return nil
	}

	return hc.Render(pages.LoginPage())
}
