package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/messaging"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/model"
	"github.com/rs/zerolog/log"
)

const gracefulShutdownDuration = time.Second * 10

func main() {
	go func() {
		_ = messaging.DefaultEventBus.ListenAndServe()
	}()
	defer func() {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), gracefulShutdownDuration)
		_ = messaging.DefaultEventBus.Shutdown(ctxTimeout)
		cancel()
	}()
	lambda.Start(Handler)
}

func Handler(ctx context.Context, events events.DynamoDBEvent) error {
	for _, event := range events.Records {
		if event.EventName == "INSERT" {
			txMessage := event.Change.NewImage
			msgBody, ok := txMessage["events"]
			if !ok {
				log.Print(errors.New("handler: events field is nil").Error())
				continue
			}
			var eventItems []model.EventItem
			if err := json.Unmarshal([]byte(msgBody.String()), &eventItems); err != nil {
				log.Print(err.Error())
				continue
			}
			log.Printf("%+v", eventItems)
			for _, eventItem := range eventItems {
				err := messaging.DefaultEventBus.PublishWithTopic(ctx, eventItem.Topic,
					eventItem.DomainEvent)
				if err != nil {
					continue
				}
			}
		}
	}
	return nil
}
