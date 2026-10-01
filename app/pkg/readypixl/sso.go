// Package readypixl holds the ReadyPixl-specific additions to this Fider fork.
//
// Single sign-on: the board has no login of its own. The ReadyPixl app is the
// identity authority. When an anonymous browser opens the board, it is bounced
// once to READYPIXL_SSO_URL with prompt=none; if that browser is signed in to
// ReadyPixl, the app sends it back to /sso/readypixl with a short-lived HS256
// token (signed with READYPIXL_SSO_SECRET) and the board signs the user in.
package readypixl

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/getfider/fider/app/pkg/web"
)

// Provider is the provider name stored on board users who signed in through ReadyPixl.
const Provider = "readypixl"

// Audience is the required "aud" claim of a ReadyPixl sign-in token.
const Audience = "readypixl-feedback"

// CheckedCookie marks a browser whose ReadyPixl session was already checked,
// so the silent check runs at most once per window.
const CheckedCookie = "rp_sso_checked"

// MaxTokenLifetime is the longest a sign-in token may live.
const MaxTokenLifetime = 10 * time.Minute

// Secret returns the shared signing secret, or "" when SSO is not configured.
func Secret() string {
	return os.Getenv("READYPIXL_SSO_SECRET")
}

// Enabled reports whether ReadyPixl single sign-on is configured.
func Enabled() bool {
	return Secret() != "" && os.Getenv("READYPIXL_SSO_URL") != ""
}

// IsAdminEmail reports whether email is listed in READYPIXL_ADMIN_EMAILS (comma separated).
func IsAdminEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, admin := range strings.Split(os.Getenv("READYPIXL_ADMIN_EMAILS"), ",") {
		if strings.ToLower(strings.TrimSpace(admin)) == email {
			return true
		}
	}
	return false
}

// StartURL is the ReadyPixl app address that checks the user's session and
// returns them to the board. prompt is "none" (silent check) or "login".
func StartURL(baseURL, redirect, prompt string) string {
	returnTo := baseURL + "/sso/readypixl?redirect=" + url.QueryEscape(SafeRedirect(redirect))
	start, err := url.Parse(os.Getenv("READYPIXL_SSO_URL"))
	if err != nil {
		return baseURL + SafeRedirect(redirect)
	}
	q := start.Query()
	q.Set("prompt", prompt)
	q.Set("return_to", returnTo)
	start.RawQuery = q.Encode()
	return start.String()
}

// SafeRedirect only allows paths on this board, so the sign-in flow cannot be
// used to send people to another site.
func SafeRedirect(redirect string) string {
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") || strings.HasPrefix(redirect, "/\\") {
		return "/"
	}
	return redirect
}

// MarkChecked sets the cookie that pauses the silent session check for d.
func MarkChecked(c *web.Context, d time.Duration) {
	http.SetCookie(&c.Response, &http.Cookie{
		Name:     CheckedCookie,
		Value:    "1",
		Path:     "/",
		Expires:  time.Now().Add(d),
		HttpOnly: true,
		Secure:   c.Request.IsSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

var skipPrefixes = []string{
	"/sso/", "/signin", "/signout", "/signup", "/oauth/", "/static/", "/assets/",
	"/_api/", "/api/", "/feed/", "/llms", "/sitemap.xml", "/robots.txt",
	"/privacy", "/terms", "/_design", "/favicon",
}

var nonBrowserRegex = regexp.MustCompile(`(?i)(bot|crawl|spider|slurp|preview|facebookexternalhit|curl|wget|python|go-http-client|node-fetch|axios|httpclient)`)

// ShouldSilentCheck reports whether this request should be bounced to
// ReadyPixl to pick up an existing session.
func ShouldSilentCheck(c *web.Context) bool {
	if !Enabled() || c.Request.Method != http.MethodGet || c.User() != nil {
		return false
	}
	path := c.Request.URL.Path
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}
	if !strings.Contains(c.Request.GetHeader("Accept"), "text/html") {
		return false
	}
	ua := c.Request.GetHeader("User-Agent")
	if ua == "" || c.Request.IsCrawler() || nonBrowserRegex.MatchString(ua) {
		return false
	}
	if cookie, err := c.Request.Cookie(CheckedCookie); err == nil && cookie.Value != "" {
		return false
	}
	return true
}
