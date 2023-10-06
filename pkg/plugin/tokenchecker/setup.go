package tokenchecker

import (
	"github.com/hellofresh/janus/pkg/config"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

func init() {
	plugin.RegisterPlugin("tokenchecker", plugin.Plugin{
		Action:   setupTokenChecker,
		Validate: nil,
	})
}

func setupTokenChecker(def *proxy.RouterDefinition, _ plugin.Config) error {
	var conf config.Config

	err := config.UnmarshalYAML("./config/config.yaml", &conf)
	if err != nil {
		return err
	}
	conf.KafkaConfig.Normalize()

	//TODO: made this works with only one instance of token array, only update it when proper fact received via kafka
	//var tokens []*models.JWTToken
	//
	//err = service.UpdateTokens(&conf, &tokens)
	//if err != nil {
	//	return err
	//}
	//
	//go service.StartFactConsumer(&conf, nil, &tokens)

	def.AddMiddleware(NewTokenCheckerMiddleware(&conf))

	return nil
}
