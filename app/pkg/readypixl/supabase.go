package readypixl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// SupabaseURL is ReadyPixl's Supabase project URL, the same public value readypixl.com ships.
func SupabaseURL() string {
	return strings.TrimRight(os.Getenv("READYPIXL_SUPABASE_URL"), "/")
}

// SupabaseAnonKey is ReadyPixl's public (anon) Supabase key, the same value readypixl.com ships.
func SupabaseAnonKey() string {
	return os.Getenv("READYPIXL_SUPABASE_ANON_KEY")
}

// SupabaseUser is the part of a ReadyPixl account the board needs.
type SupabaseUser struct {
	ID               string         `json:"id"`
	Email            string         `json:"email"`
	EmailConfirmedAt string         `json:"email_confirmed_at"`
	UserMetadata     map[string]any `json:"user_metadata"`
}

// DisplayName is the account's name, if ReadyPixl knows one.
func (u *SupabaseUser) DisplayName() string {
	for _, key := range []string{"full_name", "name"} {
		if name, ok := u.UserMetadata[key].(string); ok && strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

var supabaseClient = &http.Client{Timeout: 10 * time.Second}

// GetSupabaseUser asks ReadyPixl's Supabase who owns accessToken. The board trusts
// only Supabase's answer, never what the browser claims.
func GetSupabaseUser(ctx context.Context, accessToken string) (*SupabaseUser, error) {
	if SupabaseURL() == "" || SupabaseAnonKey() == "" {
		return nil, errors.New("ReadyPixl sign-in is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, SupabaseURL()+"/auth/v1/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", SupabaseAnonKey())
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := supabaseClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ReadyPixl rejected the session (HTTP %d)", res.StatusCode)
	}

	user := &SupabaseUser{}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(user); err != nil {
		return nil, err
	}
	if user.ID == "" || strings.TrimSpace(user.Email) == "" {
		return nil, errors.New("ReadyPixl account has no id or email")
	}
	return user, nil
}
