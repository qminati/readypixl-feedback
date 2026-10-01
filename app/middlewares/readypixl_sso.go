package middlewares

import (
	"time"

	"github.com/getfider/fider/app/pkg/readypixl"
	"github.com/getfider/fider/app/pkg/web"
)

// ReadyPixlSilentSignIn bounces an anonymous browser once to the ReadyPixl app,
// which returns it signed in when it already has a ReadyPixl session.
func ReadyPixlSilentSignIn() web.MiddlewareFunc {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(c *web.Context) error {
			if readypixl.ShouldSilentCheck(c) {
				readypixl.MarkChecked(c, 30*time.Minute)
				return c.Redirect(readypixl.StartURL(c.BaseURL(), c.Request.URL.RequestURI(), "none"))
			}
			return next(c)
		}
	}
}
