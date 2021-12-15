package persistence

import (
	"context"
	"sync"

	"github.com/maestre3d/dynamodb-tx-outbox/domain/aggregate"
	"github.com/maestre3d/dynamodb-tx-outbox/domain/repository"
)

type StudentInMemory struct {
	db map[string]aggregate.Student
	mu sync.RWMutex
}

var _ repository.Student = &StudentInMemory{}

func NewStudentInMemory() *StudentInMemory {
	return &StudentInMemory{
		db: map[string]aggregate.Student{},
		mu: sync.RWMutex{},
	}
}

func (s *StudentInMemory) Save(_ context.Context, student *aggregate.Student) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db[student.StudentID] = *student
	return nil
}

func (s *StudentInMemory) Fetch(_ context.Context, _, studentID string) (*aggregate.Student, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	student, ok := s.db[studentID]
	if !ok {
		return nil, nil
	}
	return &student, nil
}
