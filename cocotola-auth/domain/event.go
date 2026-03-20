package domain

import "time"

// EventTypeAppUserCreated is the event type identifier for AppUserCreated.
const EventTypeAppUserCreated = "AppUserCreated"

// Event represents a domain event that occurred within an aggregate.
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// EventPublisher publishes domain events asynchronously.
type EventPublisher interface {
	Publish(event Event)
}

// AppUserCreated is emitted when a new app user is created.
type AppUserCreated struct {
	AppUserID      int
	OrganizationID int
	LoginID        string
	occurredAt     time.Time
}

// NewAppUserCreated returns a new AppUserCreated event.
func NewAppUserCreated(appUserID int, organizationID int, loginID string, occurredAt time.Time) AppUserCreated {
	return AppUserCreated{
		AppUserID:      appUserID,
		OrganizationID: organizationID,
		LoginID:        loginID,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e AppUserCreated) EventType() string { return EventTypeAppUserCreated }

// OccurredAt returns the time the event occurred.
func (e AppUserCreated) OccurredAt() time.Time { return e.occurredAt }
