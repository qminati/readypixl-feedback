package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/pkg/readypixl"
	"github.com/getfider/fider/app/pkg/web"
	webutil "github.com/getfider/fider/app/pkg/web/util"
)

// ReadyPixlSignInPage shows the same sign-in card as readypixl.com/sign-in, backed by
// ReadyPixl's own accounts (its Supabase project), so the board login is the ReadyPixl login.
func ReadyPixlSignInPage() web.HandlerFunc {
	return func(c *web.Context) error {
		if c.User() != nil {
			return c.Redirect(c.BaseURL() + readypixl.SafeRedirect(c.QueryParam("redirect")))
		}

		return c.Page(http.StatusOK, web.Props{
			Page:  "ReadyPixlSignIn/ReadyPixlSignIn.page",
			Title: "Sign in to ReadyPixl",
			Data: web.Map{
				"supabaseUrl":     readypixl.SupabaseURL(),
				"supabaseAnonKey": readypixl.SupabaseAnonKey(),
			},
		})
	}
}

type readyPixlSessionInput struct {
	AccessToken string `json:"accessToken"`
}

// ReadyPixlSession signs the browser in to the board with a ReadyPixl session, once
// ReadyPixl's Supabase confirms the session and the account's email is confirmed.
func ReadyPixlSession() web.HandlerFunc {
	return func(c *web.Context) error {
		input := readyPixlSessionInput{}
		if err := json.Unmarshal([]byte(c.Request.Body), &input); err != nil || input.AccessToken == "" {
			return c.BadRequest(web.Map{})
		}

		account, err := readypixl.GetSupabaseUser(c, input.AccessToken)
		if err != nil {
			log.Warnf(c, "ReadyPixl sign-in refused: @{Error}", dto.Props{"Error": err.Error()})
			return c.Unauthorized()
		}
		if account.EmailConfirmedAt == "" {
			return c.Unauthorized()
		}

		claims := &readyPixlClaims{Email: account.Email, EmailVerified: true, Name: account.DisplayName()}
		claims.Subject = account.ID

		user, err := findOrCreateReadyPixlUser(c, claims)
		if err != nil {
			return c.Failure(err)
		}
		if user.Status == enum.UserBlocked {
			return c.Unauthorized()
		}

		webutil.AddAuthUserCookie(c, user)
		return c.Ok(web.Map{})
	}
}
