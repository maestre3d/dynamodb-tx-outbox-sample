package aggregate

import "github.com/maestre3d/dynamodb-tx-outbox/domain/event"

type Root interface {
	PullDomainEvents() []event.Domain
}

type root struct {
	events []event.Domain
}

var _ Root = &root{}

func (r *root) pushDomainEvents(events ...event.Domain) {
	r.events = append(r.events, events...)
}

func (r *root) PullDomainEvents() []event.Domain {
	memoizedEvents := r.events
	r.events = make([]event.Domain, 0)
	return memoizedEvents
}
