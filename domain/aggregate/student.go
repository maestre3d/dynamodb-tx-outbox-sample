package aggregate

import "github.com/maestre3d/dynamodb-tx-outbox/domain/event"

type Student struct {
	*root

	StudentID string
	Name      string
	SchoolID  string
}

func NewStudent(id, name, schoolID string) *Student {
	s := &Student{
		root:      &root{events: []event.Domain{}},
		StudentID: id,
		Name:      name,
		SchoolID:  schoolID,
	}
	s.pushDomainEvents(event.StudentRegistered{
		StudentID: s.StudentID,
		Name:      s.Name,
		SchoolID:  s.SchoolID,
	})
	return s
}
