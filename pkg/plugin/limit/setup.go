package authorization

import (
	"strconv"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

const maxRequestSizeInMBField = "max_request_size_in_mb"

var defaultMaxRequestSizeInMB float64

func init() {
	plugin.RegisterEventHook(plugin.StartupEvent, onStartup)
	plugin.RegisterPlugin("limit", plugin.Plugin{
		Action:   setupLimit,
		Validate: nil,
	})
}

func onStartup(event interface{}) error {
	startupEvent, ok := event.(plugin.OnStartup)
	if !ok {
		return ErrEventTypeConvert
	}

	defaultMaxRequestSizeInMB = startupEvent.Config.Limit.DefaultMaxRequestSizeInMB

	return nil
}

func setupLimit(def *proxy.RouterDefinition, cfg plugin.Config) error {
	var maxRequestSizeInMB float64

	uniqueMaxRequestSizeInMB, exists := cfg[maxRequestSizeInMBField]
	if exists {
		switch value := uniqueMaxRequestSizeInMB.(type) {
		case float64:
			maxRequestSizeInMB = value
		case string:
			parsedMaxRequestSizeInMB, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			maxRequestSizeInMB = parsedMaxRequestSizeInMB
		}
	} else {
		if defaultMaxRequestSizeInMB == 0 {
			return ErrNoDefaultLimitSize
		}
		maxRequestSizeInMB = defaultMaxRequestSizeInMB
	}

	def.AddMiddleware(NewSizeLimitMiddleware(maxRequestSizeInMB))

	return nil
}
