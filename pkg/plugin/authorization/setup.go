package authorization

import (
	"time"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

var (
	tm *TokenManager
	rm *RoleManager
)

const (
	endpointTypeField = "endpoint_type"
	loginType         = "login"
	logoutType        = "logout"

	retryAttempts = 20
	retryTimeout  = 3 * time.Second
)

func init() {
	plugin.RegisterEventHook(plugin.StartupEvent, onStartup)
	plugin.RegisterPlugin("authorization", plugin.Plugin{
		Action:   setupAuthorization,
		Validate: nil,
	})
}

func onStartup(event interface{}) error {
	startupEvent, ok := event.(plugin.OnStartup)
	if !ok {
		return ErrEventTypeConvert
	}

	tm = NewTokenManager(&startupEvent.Config.Authorization)
	rm = NewRoleManager(&startupEvent.Config.Authorization)

	err := tm.FetchTokensWithRetry(retryAttempts, retryTimeout)
	if err != nil {
		return err
	}
	err = rm.FetchRolesWithRetry(retryAttempts, retryTimeout)
	if err != nil {
		return err
	}

	return nil
}

func setupAuthorization(def *proxy.RouterDefinition, cfg plugin.Config) error {
	endpointType, exists := cfg[endpointTypeField]
	if exists {
		switch endpointType {
		case loginType:
			def.AddMiddleware(NewLoginTokenCatcherMiddleware(tm))
		case logoutType:
			def.AddMiddleware(NewLogoutTokenCatcherMiddleware(tm))
		}
		return nil
	}

	def.AddMiddleware(NewTokenCheckerMiddleware(tm))
	def.AddMiddleware(NewRoleCheckerMiddleware(rm))

	return nil
}
