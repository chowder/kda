package server

import (
	"crypto/tls"
	"fmt"
	"github.com/chowder/kda/tokensource"
	"github.com/chowder/kda/validator"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const AuthTokenCookieName = "auth_token"

type Server struct {
	tokenSource tokensource.TokenSource
	validator   validator.Validator
}

func NewServer(tokenSource tokensource.TokenSource, validator validator.Validator) *Server {
	return &Server{
		tokenSource: tokenSource,
		validator:   validator,
	}
}

func (s *Server) Serve(addr string, backend *url.URL) error {
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check for the authentication token cookie
		cookie, err := r.Cookie(AuthTokenCookieName)
		if err == nil {
			// Token is present, forward the request to the backend service
			r.Header.Set("Authorization", "Bearer "+cookie.Value)
			proxy.ServeHTTP(w, r)
			return
		}

		// Check for the basic auth header
		user, pass, ok := r.BasicAuth()
		if !ok {
			// If not present, send a challenge
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("401 Unauthorized\n"))
			return
		}

		if s.validator.Validate(user, pass) {
			token, err := s.tokenSource.Create(r.Context())
			if err != nil {
				log.Println("error creating service account token: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("500 Internal Server Error\n"))
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  AuthTokenCookieName,
				Value: token.Value,
				Path:  "/",
				// Make the cookie expire just a little bit before the actual token expiry
				Expires:  token.ExpiresAt.Add(-1 * time.Minute),
				HttpOnly: true,
			})

			proxy.ServeHTTP(w, r)
			return
		} else {
			log.Println("unable to validate user: ", user)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("401 Unauthorized\n"))
			return
		}
	})

	// Start the reverse proxy server
	log.Println("Starting reverse proxy server on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("could not start reverse proxy server: %w", err)
	}

	return nil
}
