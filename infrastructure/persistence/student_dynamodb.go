package persistence

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/repository"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/model"
)

const (
	StudentDynamoDbTableName = "students"
	OutboxDynamoDbTableName  = "outbox"
)

type StudentDynamoDb struct {
	db *dynamodb.Client
}

var _ repository.Student = &StudentDynamoDb{}

func NewStudentDynamoDb(db *dynamodb.Client) *StudentDynamoDb {
	return &StudentDynamoDb{
		db: db,
	}
}

func (s *StudentDynamoDb) Save(ctx context.Context, student *aggregate.Student) error {
	txItems := make([]types.TransactWriteItem, 0)
	txItems = append(txItems, types.TransactWriteItem{
		Put: &types.Put{
			Item:      model.MarshalStudentDynamoDb(student),
			TableName: aws.String(StudentDynamoDbTableName),
		},
	})
	txItems = appendDomainEventsIntoTx(txItems, student)
	_, err := s.db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: txItems,
	})
	return err
}

func (s *StudentDynamoDb) Fetch(ctx context.Context, schoolID, studentID string) (*aggregate.Student, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"student_id": &types.AttributeValueMemberS{Value: studentID},
			"school_id":  &types.AttributeValueMemberS{Value: schoolID},
		},
		TableName: aws.String(StudentDynamoDbTableName),
	})
	if err != nil {
		return nil, err
	}

	return model.UnmarshalStudentDynamoDb(out.Item), nil
}
