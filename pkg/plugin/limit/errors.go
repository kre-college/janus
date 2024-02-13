package authorization

import (
	"errors"
	"net/http"

	errorsJanus "github.com/hellofresh/janus/pkg/errors"
)

var (
	ErrRequestTooLarge = errorsJanus.New(http.StatusRequestEntityTooLarge, "request too large")
)

var (
	ErrEventTypeConvert   = errors.New("cannot convert event")
	ErrNoDefaultLimitSize = errors.New("cannot find default request limit size")
)
