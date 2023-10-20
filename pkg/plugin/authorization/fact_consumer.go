package authorization

import (
	"encoding/json"

	"github.com/hellofresh/janus/pkg/config"
	"github.com/hellofresh/janus/pkg/kafka"
	"github.com/hellofresh/janus/pkg/models"
)

func StartFactConsumers(conf *config.Config, tokenManager *models.TokenManager, roleManager *models.RoleManager) {
	for _, topic := range conf.KafkaConfig.Consumer.Topics {
		switch topic {

		case "user-management-facts":
			for i := 0; i < conf.KafkaConfig.Consumer.ConsumersAmount; i++ {
				go UserManagementFactConsumer(conf, topic, tokenManager)
			}
		case "roles":
			for i := 0; i < conf.KafkaConfig.Consumer.ConsumersAmount; i++ {
				go RBACFactConsumer(conf, topic, roleManager)
			}
		}
	}
}

func UserManagementFactConsumer(conf *config.Config, topic string, tokenManager *models.TokenManager) {
	kafka.ConsumeMessages(conf.KafkaConfig, topic,
		func(msg kafka.Message) error {
			var fact models.Fact

			err := json.Unmarshal(msg.Value, &fact)
			if err != nil {
				return err
			}

			switch fact.ObjectType {

			case models.ObjectTypeJWTToken:
				switch fact.ActionType {

				case models.ActionTypeCreate, models.ActionTypeUpdate:
					var token *models.JWTToken
					err = json.Unmarshal(*fact.Object, &token)
					if err != nil {
						return err
					}
					err = UpsertToken(token, tokenManager)
					if err != nil {
						return err
					}

				case models.ActionTypeDelete:
					err = DeleteTokenByID(fact.ID, tokenManager)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		func(msg kafka.Message, inerr error) {
			msg.Headers = []kafka.Header{{Key: "Error", Value: []byte(inerr.Error())}}
			producer := kafka.NewKafkaProducer(conf.KafkaConfig)
			defer producer.Close()
			producer.ProduceMadeMessage(msg)
			return
		},
	)
}

func RBACFactConsumer(conf *config.Config, topic string, roleManager *models.RoleManager) {
	kafka.ConsumeMessages(conf.KafkaConfig, topic,
		func(msg kafka.Message) error {
			var fact models.Fact

			err := json.Unmarshal(msg.Value, &fact)
			if err != nil {
				return err
			}

			switch fact.ObjectType {

			case models.ObjectTypeRole:
				switch fact.ActionType {

				case models.ActionTypeCreate, models.ActionTypeUpdate:
					var role *models.Role
					err = json.Unmarshal(*fact.Object, &role)
					if err != nil {
						return err
					}

					err = UpsertRole(role, roleManager)
					if err != nil {
						return err
					}
				}
			case models.ActionTypeDelete:
				err = DeleteRoleByID(fact.ID, roleManager)
				if err != nil {
					return err
				}
			}
			return nil
		},
		func(msg kafka.Message, inerr error) {
			msg.Headers = []kafka.Header{{Key: "Error", Value: []byte(inerr.Error())}}
			producer := kafka.NewKafkaProducer(conf.KafkaConfig)
			defer producer.Close()
			producer.ProduceMadeMessage(msg)
			return
		},
	)
}