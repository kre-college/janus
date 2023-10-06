package tokenchecker

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

func NewTokenCheckerMiddleware(conf *config.Config) func(http.Handler) http.Handler {
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

			var tokens []*models.JWTToken

			err := service.UpdateTokens(conf, &tokens)
			if err != nil {
				errors.Handler(w, r, errors.New(http.StatusInternalServerError, err.Error()))
				return
			}

			err = TokenChecker(tokens, accessToken)
			if err != nil {
				errors.Handler(w, r, errors.New(http.StatusUnauthorized, err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), "auth_header", accessToken)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TokenChecker(tokens []*models.JWTToken, userToken string) error {
	for _, token := range tokens {
		if token.Token == userToken {
			return nil
		}
	}

	return fmt.Errorf("invalid token")
}
