package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
)

func MarshalStudentDynamoDb(student *aggregate.Student) map[string]types.AttributeValue {
	if student == nil {
		return nil
	}
	return map[string]types.AttributeValue{
		"student_id":   &types.AttributeValueMemberS{Value: student.StudentID},
		"student_name": &types.AttributeValueMemberS{Value: student.Name},
		"school_id":    &types.AttributeValueMemberS{Value: student.SchoolID},
	}
}

func UnmarshalStudentDynamoDb(studentDynamo map[string]types.AttributeValue) *aggregate.Student {
	if studentDynamo == nil {
		return nil
	}
	return &aggregate.Student{
		StudentID: studentDynamo["student_id"].(*types.AttributeValueMemberS).Value,
		Name:      studentDynamo["student_name"].(*types.AttributeValueMemberS).Value,
		SchoolID:  studentDynamo["school_id"].(*types.AttributeValueMemberS).Value,
	}
}
