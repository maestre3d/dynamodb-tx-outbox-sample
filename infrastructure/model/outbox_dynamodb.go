package model

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
)

func MarshalOutboxDynamoDb(ag aggregate.Root) map[string]types.AttributeValue {
	// Write to outbox in transaction batches.
	// Thus, every publish job from the log trailing daemon will publish messages in a batch mode per-tx and network usage
	// will be reduced
	eventsJSON, err := json.Marshal(ag.PullDomainEvents())
	if err != nil {
		return nil
	}
	return map[string]types.AttributeValue{
		"transaction_id": &types.AttributeValueMemberS{
			Value: uuid.NewString(),
		},
		"message_body": &types.AttributeValueMemberB{
			Value: eventsJSON,
		},
		"occurred_at": &types.AttributeValueMemberS{
			Value: time.Now().UTC().Format(time.RFC3339),
		},
		"time_to_exist": &types.AttributeValueMemberS{
			Value: strconv.Itoa(int(time.Now().Add(time.Minute * 60 * 24).Unix())),
		},
	}
}
