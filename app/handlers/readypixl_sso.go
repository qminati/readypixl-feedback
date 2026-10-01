package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/log"
	"github.com/getfider/fider/app/pkg/readypixl"
	"github.com/getfider/fider/app/pkg/web"
	webutil "github.com/getfider/fider/app/pkg/web/util"
	jwtgo "github.com/golang-jwt/jwt/v4"
)

// readyPixlClaims is the sign-in token the ReadyPixl app issues for its signed-in user.
type readyPixlClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	// Team is set by the ReadyPixl app for its team accounts (server-only app_metadata.developer).
	// Not used to grant any role yet: board admins come only from READYPIXL_ADMIN_EMAILS.
	Team bool `json:"team"`
	jwtgo.RegisteredClaims
}

// ReadyPixlSignInStart sends the browser to the ReadyPixl login, which returns it here signed in.
func ReadyPixlSignInStart() web.HandlerFunc {
	return func(c *web.Context) error {
		redirect := readypixl.SafeRedirect(c.QueryParam("redirect"))
		if !readypixl.Enabled() {
			return c.Redirect("/signin")
		}
		return c.Redirect(readypixl.StartURL(c.BaseURL(), redirect, "login"))
	}
}

// ReadyPixlSSO receives the browser back from the ReadyPixl app, with a sign-in token
// when the browser is signed in to ReadyPixl, and without one when it is not.
func ReadyPixlSSO() web.HandlerFunc {
	return func(c *web.Context) error {
		c.Response.Header().Add("X-Robots-Tag", "noindex")
		redirect := readypixl.SafeRedirect(c.QueryParam("redirect"))
		readypixl.MarkChecked(c, 30*time.Minute)

		// Accepting a pass only needs the shared secret; Enabled() also needs the
		// ReadyPixl address, which only the outgoing bounce uses.
		token := c.QueryParam("token")
		if token == "" || readypixl.Secret() == "" {
			return c.Redirect(redirect)
		}

		claims, err := decodeReadyPixlToken(token)
		if err != nil {
			log.Warnf(c, "Rejected ReadyPixl sign-in token: @{Error}", dto.Props{"Error": err.Error()})
			return c.Redirect(redirect)
		}

		user, err := findOrCreateReadyPixlUser(c, claims)
		if err != nil {
			return c.Failure(err)
		}
		if user.Status == enum.UserBlocked {
			return c.Redirect(redirect)
		}

		webutil.AddAuthUserCookie(c, user)
		return c.Redirect(redirect)
	}
}

func decodeReadyPixlToken(token string) (*readyPixlClaims, error) {
	claims := &readyPixlClaims{}
	parsed, err := jwtgo.ParseWithClaims(token, claims, func(t *jwtgo.Token) (any, error) {
		if t.Method != jwtgo.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(readypixl.Secret()), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	if !claims.VerifyAudience(readypixl.Audience, true) {
		return nil, fmt.Errorf("wrong audience")
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) > readypixl.MaxTokenLifetime {
		return nil, fmt.Errorf("token lifetime missing or too long")
	}
	if claims.Subject == "" || strings.TrimSpace(claims.Email) == "" {
		return nil, fmt.Errorf("token has no user id or email")
	}
	// Only confirmed emails: otherwise anyone could register on ReadyPixl with
	// someone else's address and take over their board account.
	if !claims.EmailVerified {
		return nil, fmt.Errorf("email is not verified")
	}
	return claims, nil
}

func findOrCreateReadyPixlUser(c *web.Context, claims *readyPixlClaims) (*entity.User, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	name := strings.TrimSpace(claims.Name)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	byProvider := &query.GetUserByProvider{Provider: readypixl.Provider, UID: claims.Subject}
	err := bus.Dispatch(c, byProvider)
	user := byProvider.Result

	if errors.Cause(err) == app.ErrNotFound {
		byEmail := &query.GetUserByEmail{Email: email}
		err = bus.Dispatch(c, byEmail)
		user = byEmail.Result
		if err == nil && !user.HasProvider(readypixl.Provider) {
			if err := bus.Dispatch(c, &cmd.RegisterUserProvider{
				UserID:       user.ID,
				ProviderName: readypixl.Provider,
				ProviderUID:  claims.Subject,
			}); err != nil {
				return nil, err
			}
		}
	}

	if errors.Cause(err) == app.ErrNotFound {
		user = &entity.User{
			Name:   name,
			Tenant: c.Tenant(),
			Email:  email,
			Role:   enum.RoleVisitor,
			Providers: []*entity.UserProvider{
				{UID: claims.Subject, Name: readypixl.Provider},
			},
		}
		if err := bus.Dispatch(c, &cmd.RegisterUser{User: user}); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if readypixl.IsAdminEmail(email) && user.Role != enum.RoleAdministrator {
		if err := bus.Dispatch(c, &cmd.ChangeUserRole{UserID: user.ID, Role: enum.RoleAdministrator}); err != nil {
			return nil, err
		}
		user.Role = enum.RoleAdministrator
	}

	return user, nil
}
