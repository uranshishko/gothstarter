package handlers

import (
	"encoding/json"

	"github.com/uranshishko/gothstarter/auth"
	"github.com/uranshishko/gothstarter/common"
	"github.com/uranshishko/gothstarter/views/pages"
)

func SettingsHandler(hc common.HandlerContext) error {
	userStr, err := auth.Client.GetFromSession(hc, "user")
	if err != nil {
		hc.Redirect(302, "/login")
		return nil
	}

	var user auth.User
	err = json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		hc.Redirect(302, "/login")
		return nil
	}

	return hc.Render(pages.SettingsPage(user))
}
