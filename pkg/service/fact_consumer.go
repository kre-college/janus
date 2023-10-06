package service

import (
	"context"
	"encoding/json"

	"github.com/hellofresh/janus/pkg/config"
	pg "github.com/hellofresh/janus/pkg/db/postgres"
	"github.com/hellofresh/janus/pkg/kafka"
	"github.com/hellofresh/janus/pkg/models"
)

func StartFactConsumer(conf *config.Config, roles *[]*models.Role, tokens *[]*models.JWTToken) {
	for _, topic := range conf.KafkaConfig.Consumer.Topics {
		switch topic {
		case "rbac-facts":
			db, err := pg.NewDB(conf.DBRbac.URL())
			if err != nil {
				return
			}
			for i := 0; i < conf.KafkaConfig.Consumer.ConsumersAmount; i++ {
				go StartRBACFactConsumer(conf.KafkaConfig, db, topic, roles)
			}
		case "user-management-facts":
			db, err := pg.NewDB(conf.DBUserManagement.URL())
			if err != nil {
				return
			}
			for i := 0; i < conf.KafkaConfig.Consumer.ConsumersAmount; i++ {
				go StartUserManagementFactConsumer(conf.KafkaConfig, db, topic, tokens)
			}
		}
	}
}

func StartRBACFactConsumer(conf *kafka.Config, db *pg.DB, topic string, roles *[]*models.Role) {
	kafka.ConsumeMessages(conf, topic,
		func(msg kafka.Message) error {
			var fact models.Fact

			ctx := context.Background()

			err := json.Unmarshal(msg.Value, &fact)
			if err != nil {
				return err
			}

			switch fact.ObjectType {

			case models.ObjectTypeRole:
				switch fact.ActionType {

				case models.ActionTypeCreate:
				case models.ActionTypeUpdate:
				case models.ActionTypeDelete:
					err = fetchRoles(ctx, db, roles)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		func(msg kafka.Message, inerr error) {
			msg.Headers = []kafka.Header{{Key: "Error", Value: []byte(inerr.Error())}}
			producer := kafka.NewKafkaProducer(conf)
			defer producer.Close()
			producer.ProduceMadeMessage(msg)
			return
		},
	)
}

func StartUserManagementFactConsumer(conf *kafka.Config, db *pg.DB, topic string, tokens *[]*models.JWTToken) {
	kafka.ConsumeMessages(conf, topic,
		func(msg kafka.Message) error {
			var fact models.Fact

			ctx := context.Background()

			err := json.Unmarshal(msg.Value, &fact)
			if err != nil {
				return err
			}

			switch fact.ObjectType {

			case models.ObjectTypeJWTToken:
				switch fact.ActionType {

				case models.ActionTypeCreate:
				case models.ActionTypeUpdate:
				case models.ActionTypeDelete:
					err = fetchTokens(ctx, db, tokens)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		func(msg kafka.Message, inerr error) {
			msg.Headers = []kafka.Header{{Key: "Error", Value: []byte(inerr.Error())}}
			producer := kafka.NewKafkaProducer(conf)
			defer producer.Close()
			producer.ProduceMadeMessage(msg)
			return
		},
	)
}
