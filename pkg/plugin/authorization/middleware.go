package authorization

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/hellofresh/janus/pkg/errors"
	"io"
)

const (
	AuthHeaderKey = "Authorization"
)

var (
	AuthHeaderValue = ContextKey("auth_header")
)

// ContextKey is used to create context keys that are concurrent safe
type ContextKey string

func (c ContextKey) String() string {
	return "janus." + string(c)
}

func getAccessToken(r *http.Request) (string, error) {
	authHeaderValue := r.Header.Get(AuthHeaderKey)
	parts := strings.Split(authHeaderValue, " ")

	if len(parts) == 0 {
		return "", ErrAuthorizationFieldNotFound
	} else if len(parts) < 2 {
		logrus.Errorf("bearer token malformed, token is: %q", authHeaderValue)
		return "", ErrBearerMalformed
	}

	accessToken := parts[1]

	return accessToken, nil
}

func NewTokenCheckerMiddleware(manager *TokenManager) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken, err := getAccessToken(r)
			if err != nil {
				errors.Handler(w, r, err)
				return
			}
			if !manager.isTokenValid(accessToken) {
				errors.Handler(w, r, ErrAccessTokenNotAuthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AuthHeaderValue, accessToken)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewLoginTokenCatcherMiddleware(manager *TokenManager) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcw := &responseCatcherWriter{ResponseWriter: w}

			tm.mu.RLock()
			defer tm.mu.RUnlock()
			handler.ServeHTTP(rcw, r)

			if rcw.status != http.StatusOK {
				return
			}

			var body []byte
			encoding := rcw.Header().Get("Content-Encoding")
			if strings.Contains(encoding, "gzip") {
				gzr, err := gzip.NewReader(bytes.NewReader(rcw.body))
				if err != nil {
					fmt.Println("error creating gzip reader")
					fmt.Println(err)
					return
				}
				defer gzr.Close()

				unzippedBody, err := io.ReadAll(gzr)
				if err != nil {
					fmt.Println("error reading unzipped body")
					fmt.Println(err)
					return
				}
				body = unzippedBody
			} else {
				body = rcw.body
			}

			var accessToken string
			err := json.Unmarshal(body, &accessToken)
			if err != nil {
				fmt.Println("error unmarshalling access token")
				fmt.Println(err)
				return
			}

			err = manager.UpsertToken(accessToken)
			if err != nil {
				fmt.Println("error upserting access token")
				fmt.Println(err)
				return
			}
		})
	}
}

func NewLogoutTokenCatcherMiddleware(manager *TokenManager) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken, err := getAccessToken(r)
			if err != nil {
				errors.Handler(w, r, err)
				return
			}

			rcw := &responseCatcherWriter{ResponseWriter: w}

			handler.ServeHTTP(rcw, r)

			if rcw.status != http.StatusOK {
				return
			}

			manager.DeleteToken(accessToken)
		})
	}
}

func NewRoleCheckerMiddleware(manager *RoleManager) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken, err := getAccessToken(r)
			if err != nil {
				errors.Handler(w, r, err)
				return
			}

			claims, err := ExtractClaims(accessToken)
			if err != nil {
				errors.Handler(w, r, err)
				return
			}

			if len(claims.Roles) == 0 {
				errors.Handler(w, r, ErrNoRolesSet)
				return
			}

			if !isHaveAccess(manager.Roles, claims.Roles.GetRoles(), r.URL.Path, r.Method) {
				errors.Handler(w, r, ErrAccessIsDenied)
				return
			}

			ctx := context.WithValue(r.Context(), AuthHeaderValue, accessToken)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isHaveAccess(roles map[string]*Role, userRoles []string, path, method string) bool {
	for _, userRole := range userRoles {
		if role, exists := roles[userRole]; exists {
			for _, feature := range role.Features {
				if feature.Method == method && isEndpointPathsEqual(path, feature.Path) {
					return true
				}
			}
		}
	}

	return false
}

func isEndpointPathsEqual(reqPath, confPath string) bool {
	reqPathArr := strings.Split(confPath, "/")
	confPathArr := strings.Split(reqPath, "/")
	if len(reqPathArr) != len(confPathArr) {
		return false
	}

	for i := range confPathArr {
		if reqPathArr[i] == "" || string(reqPathArr[i][0]) == "{" {
			continue
		}

		if reqPathArr[i] != confPathArr[i] {
			return false
		}
	}

	return true
}
