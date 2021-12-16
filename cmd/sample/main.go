package main

import (
	"context"
	"encoding/json"

	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/messaging"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/model"
	"github.com/rs/zerolog/log"
)

var dummy = "[{\"topic\":\"ncorp.workspaces.iam.1.student.registered\",\"domain_event\":{\"student_id\":\"fcf9b0c0-2426-4cc9-8149-3719667a9037\",\"student_name\":\"Copias Munoz\",\"school_id\":\"123\",\"registered_at\":\"2021-12-16T11:10:36Z\"}}]"

func main() {
	go func() {
		_ = messaging.DefaultEventBus.ListenAndServe()
	}()
	var eventItems []model.EventItem
	if err := json.Unmarshal([]byte(dummy), &eventItems); err != nil {
		log.Print(err.Error())
		return
	}
	log.Printf("%+v", eventItems)
	for _, eventItem := range eventItems {
		_ = messaging.DefaultEventBus.PublishWithTopic(context.TODO(), eventItem.Topic, eventItem.DomainEvent)
	}
}
