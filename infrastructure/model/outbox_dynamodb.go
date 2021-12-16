package model

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/event"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/messaging"
)

type EventItem struct {
	Topic       string       `json:"topic"`
	DomainEvent event.Domain `json:"domain_event"`
}

func MarshalOutboxDynamoDb(ag aggregate.Root) map[string]types.AttributeValue {
	// Write to outbox in transaction batches.
	// Thus, every publish job from the log trailing daemon will publish messages in a batch mode per-tx and network usage
	// will be reduced
	eventItems := make([]EventItem, 0)
	for _, ev := range ag.PullDomainEvents() {
		metadata, err := messaging.DefaultEventBus.GetSchemaMetadata(ev)
		if err != nil {
			continue
		}
		eventItems = append(eventItems, EventItem{
			Topic:       metadata.Topic,
			DomainEvent: ev,
		})
	}
	eventsJSON, err := json.Marshal(eventItems)
	if err != nil {
		return nil
	}
	return map[string]types.AttributeValue{
		"transaction_id": &types.AttributeValueMemberS{
			Value: uuid.NewString(),
		},
		"events": &types.AttributeValueMemberS{
			Value: string(eventsJSON),
		},
		"occurred_at": &types.AttributeValueMemberS{
			Value: time.Now().UTC().Format(time.RFC3339),
		},
		"time_to_exist": &types.AttributeValueMemberS{
			Value: strconv.Itoa(int(time.Now().Add(time.Minute * 60).Unix())),
		},
	}
}
