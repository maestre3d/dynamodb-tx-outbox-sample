package persistence

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/model"
)

func appendDomainEventsIntoTx(ref []types.TransactWriteItem, ag aggregate.Root) []types.TransactWriteItem {
	ref = append(ref, types.TransactWriteItem{
		Put: &types.Put{
			Item:      model.MarshalOutboxDynamoDb(ag),
			TableName: aws.String(OutboxDynamoDbTableName),
		},
	})
	return ref
}
