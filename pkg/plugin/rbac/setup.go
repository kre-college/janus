package rbac

import (
	"github.com/hellofresh/janus/pkg/config"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

func init() {
	plugin.RegisterPlugin("rbac", plugin.Plugin{
		Action:   setupRBAC,
		Validate: nil,
	})
}

func setupRBAC(def *proxy.RouterDefinition, a plugin.Config) error {
	var conf config.Config

	err := config.UnmarshalYAML("./config/config.yaml", &conf)
	if err != nil {
		return err
	}
	conf.KafkaConfig.Normalize()

	//TODO: made this works with only one instance of roles array, only update it when proper fact received via kafka
	//var roles []*models.Role
	//
	//err = service.UpdateRoles(&conf, &roles)
	//if err != nil {
	//	return err
	//}
	//
	//go service.StartFactConsumer(&conf, &roles, nil)

	def.AddMiddleware(NewRBACMiddleware(&conf))

	return nil
}
