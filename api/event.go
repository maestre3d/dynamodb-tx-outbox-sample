package api

import "github.com/maestre3d/dynamodb-tx-outbox/controller"

func InitEventApi() {
	controller.MapStudentEventSubscriptions()
}
