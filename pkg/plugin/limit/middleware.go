package authorization

import (
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
)

func NewSizeLimitMiddleware(maxSizeInMB float64) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isRequestFitsInLimit(r, maxSizeInMB) {
				errors.Handler(w, r, ErrRequestTooLarge)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

func isRequestFitsInLimit(r *http.Request, maxSizeInMB float64) bool {
	if r.ContentLength == -1 {
		return false
	}

	if r.ContentLength > int64(maxSizeInMB*1024*1024) {
		return false
	}

	return true
}
