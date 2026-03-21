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

// EventTypeGroupCreated is the event type identifier for GroupCreated.
const EventTypeGroupCreated = "GroupCreated"

// GroupCreated is emitted when a new group is created.
type GroupCreated struct {
	GroupID        int
	OrganizationID int
	Name           string
	occurredAt     time.Time
}

// NewGroupCreated returns a new GroupCreated event.
func NewGroupCreated(groupID int, organizationID int, name string, occurredAt time.Time) GroupCreated {
	return GroupCreated{
		GroupID:        groupID,
		OrganizationID: organizationID,
		Name:           name,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e GroupCreated) EventType() string { return EventTypeGroupCreated }

// OccurredAt returns the time the event occurred.
func (e GroupCreated) OccurredAt() time.Time { return e.occurredAt }

// EventTypeSpaceCreated is the event type identifier for SpaceCreated.
const EventTypeSpaceCreated = "SpaceCreated"

// SpaceCreated is emitted when a new space is created.
type SpaceCreated struct {
	SpaceID        int
	OrganizationID int
	OwnerID        int
	KeyName        string
	Name           string
	SpaceType      string
	occurredAt     time.Time
}

// NewSpaceCreated returns a new SpaceCreated event.
func NewSpaceCreated(spaceID int, organizationID int, ownerID int, keyName string, name string, spaceType string, occurredAt time.Time) SpaceCreated {
	return SpaceCreated{
		SpaceID:        spaceID,
		OrganizationID: organizationID,
		OwnerID:        ownerID,
		KeyName:        keyName,
		Name:           name,
		SpaceType:      spaceType,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e SpaceCreated) EventType() string { return EventTypeSpaceCreated }

// OccurredAt returns the time the event occurred.
func (e SpaceCreated) OccurredAt() time.Time { return e.occurredAt }
