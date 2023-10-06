package rbac

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hellofresh/janus/pkg/config"
	"github.com/hellofresh/janus/pkg/models"
	"github.com/hellofresh/janus/pkg/service"

	"github.com/hellofresh/janus/pkg/errors"
)

func NewRBACMiddleware(conf *config.Config) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeaderValue := r.Header.Get("Authorization")
			parts := strings.Split(authHeaderValue, " ")

			if len(parts) == 0 {
				errors.Handler(w, r, errors.New(http.StatusUnauthorized, "no authorization header provided"))
				return
			} else if len(parts) == 1 {
				errors.Handler(w, r, errors.New(http.StatusUnauthorized, "bearer token malformed"))
				return
			}

			accessToken := parts[1]

			claims, err := ExtractClaims(accessToken)
			if err != nil {
				errors.Handler(w, r, err)
				return
			}

			if len(claims.Roles) <= 0 {
				errors.Handler(w, r, fmt.Errorf("no roles have been set"))
				return
			}

			var roles []*models.Role

			err = service.UpdateRoles(conf, &roles)
			if err != nil {
				errors.Handler(w, r, errors.New(http.StatusInternalServerError, err.Error()))
				return
			}

			err = RoleChecker(roles, claims.Roles, r.URL.Path, r.Method)
			if err != nil {
				errors.Handler(w, r, errors.New(http.StatusUnauthorized, err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), "auth_header", accessToken)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleChecker(roles []*models.Role, userRoles []string, path, method string) error {
	for _, userRole := range userRoles {
		for _, role := range roles {
			if role.Name != userRole {
				continue
			}
			for _, feature := range role.Features {
				if feature.Method == method && isEndpointPathsEqual(path, feature.Path) {
					return nil
				}

			}
		}
	}

	return fmt.Errorf("access denied")
}

func isEndpointPathsEqual(reqPath, dbPath string) bool {
	reqPathArr := strings.Split(dbPath, "/")
	dbPathArr := strings.Split(reqPath, "/")
	if len(reqPathArr) != len(dbPathArr) {
		return false
	}

	for i, _ := range dbPathArr {
		if reqPathArr[i] == "" || string(reqPathArr[i][0]) == "{" {
			continue
		}

		if reqPathArr[i] != dbPathArr[i] {
			return false
		}
	}

	return true
}
