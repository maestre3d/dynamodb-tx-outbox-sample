package repository

import (
	"context"

	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
)

type Student interface {
	Save(context.Context, *aggregate.Student) error
	Fetch(ctx context.Context, schoolID, studentID string) (*aggregate.Student, error)
}
