package controller

import (
	"context"

	"github.com/maestre3d/dynamodb-tx-outbox/domain/event"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/messaging"
	"github.com/neutrinocorp/gluon"
	"github.com/rs/zerolog/log"
)

func MapStudentEventSubscriptions() {
	messaging.DefaultEventBus.Subscribe(event.StudentRegistered{}).
		Group("log.on.student_registered").
		HandlerFunc(logOnStudentRegistered)
}

func logOnStudentRegistered(_ context.Context, message *gluon.Message) error {
	studentEvent, ok := message.Data.(event.StudentRegistered)
	if !ok {
		log.Print("error parsing event")
		return nil
	}
	log.Printf("%+v", studentEvent)
	return nil
}
