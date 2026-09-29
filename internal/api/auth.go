package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
)

// TokenFile is where the API token lives, next to the database.
const TokenFile = "safora.token"

const (
	cookieName   = "safora_token"
	maxBodyBytes = 1 << 20 // 1 MiB: plenty for jobs and pasted batch scripts
)

// loadOrCreateToken returns the token stored at path, creating a random one
// (owner-only permissions) on first start.
func loadOrCreateToken(path string) (string, error) {
	if b, err := os.ReadFile(path); err == nil {
		if t := strings.TrimSpace(string(b)); t != "" {
			return t, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	if err := os.WriteFile(path, []byte(token+"\n"), 0600); err != nil {
		return "", err
	}
	return token, nil
}

// secure wraps next with Host/Origin checks, a body limit and token auth.
// The dashboard (static files) receives the token as a same-site HttpOnly
// cookie; API clients may send "Authorization: Bearer <token>" instead.
func (s *Server) secure(addr string, next http.Handler) http.Handler {
	_, port, _ := net.SplitHostPort(addr)
	allowedHosts := map[string]bool{
		"127.0.0.1:" + port: true,
		"localhost:" + port: true,
		"[::1]:" + port:     true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Blocks DNS rebinding.
		if !allowedHosts[r.Host] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Blocks cross-origin browser requests (including other localhost ports).
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			if !s.validToken(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// The Live Stream overrides this with its own type.
			w.Header().Set("Content-Type", "application/json")
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    s.token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) validToken(r *http.Request) bool {
	got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if got == "" {
		if c, err := r.Cookie(cookieName); err == nil {
			got = c.Value
		}
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
}
