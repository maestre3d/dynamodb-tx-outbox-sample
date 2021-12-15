package appservice

import (
	"context"

	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/repository"
)

var StudentRepository repository.Student

type RegisterStudentArgs struct {
	StudentID string
	Name      string
	SchoolID  string
}

func RegisterStudent(ctx context.Context, args RegisterStudentArgs) error {
	return StudentRepository.Save(ctx,
		aggregate.NewStudent(args.StudentID, args.Name, args.SchoolID))
}

func GetStudentByID(ctx context.Context, schoolID, studentID string) (*aggregate.Student, error) {
	return StudentRepository.Fetch(ctx, schoolID, studentID)
}
