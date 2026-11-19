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

type baseEvent struct {
	occurredAt time.Time
}

// OccurredAt returns the time the event occurred.
func (b baseEvent) OccurredAt() time.Time { return b.occurredAt }

// AppUserCreated is emitted when a new app user is created.
type AppUserCreated struct {
	baseEvent

	appUserID      AppUserID
	organizationID OrganizationID
	loginID        string
}

// NewAppUserCreated returns a new AppUserCreated event.
func NewAppUserCreated(appUserID AppUserID, organizationID OrganizationID, loginID string, occurredAt time.Time) AppUserCreated {
	return AppUserCreated{
		baseEvent:      baseEvent{occurredAt: occurredAt},
		appUserID:      appUserID,
		organizationID: organizationID,
		loginID:        loginID,
	}
}

// EventType returns the event type identifier.
func (e AppUserCreated) EventType() string { return EventTypeAppUserCreated }

// AppUserID returns the created user's ID.
func (e AppUserCreated) AppUserID() AppUserID { return e.appUserID }

// OrganizationID returns the organization ID.
func (e AppUserCreated) OrganizationID() OrganizationID { return e.organizationID }

// LoginID returns the login ID.
func (e AppUserCreated) LoginID() string { return e.loginID }

// EventTypeGroupCreated is the event type identifier for GroupCreated.
const EventTypeGroupCreated = "GroupCreated"

// GroupCreated is emitted when a new group is created.
type GroupCreated struct {
	baseEvent

	groupID        GroupID
	organizationID OrganizationID
	name           string
}

// NewGroupCreated returns a new GroupCreated event.
func NewGroupCreated(groupID GroupID, organizationID OrganizationID, name string, occurredAt time.Time) GroupCreated {
	return GroupCreated{
		baseEvent:      baseEvent{occurredAt: occurredAt},
		groupID:        groupID,
		organizationID: organizationID,
		name:           name,
	}
}

// EventType returns the event type identifier.
func (e GroupCreated) EventType() string { return EventTypeGroupCreated }

// GroupID returns the created group's ID.
func (e GroupCreated) GroupID() GroupID { return e.groupID }

// OrganizationID returns the organization ID.
func (e GroupCreated) OrganizationID() OrganizationID { return e.organizationID }

// Name returns the group name.
func (e GroupCreated) Name() string { return e.name }

// EventTypeSpaceCreated is the event type identifier for SpaceCreated.
const EventTypeSpaceCreated = "SpaceCreated"

// SpaceCreated is emitted when a new space is created.
type SpaceCreated struct {
	baseEvent

	spaceID        SpaceID
	organizationID OrganizationID
	ownerID        AppUserID
	keyName        string
	name           string
	spaceType      string
}

// NewSpaceCreated returns a new SpaceCreated event.
func NewSpaceCreated(spaceID SpaceID, organizationID OrganizationID, ownerID AppUserID, keyName string, name string, spaceType string, occurredAt time.Time) SpaceCreated {
	return SpaceCreated{
		baseEvent:      baseEvent{occurredAt: occurredAt},
		spaceID:        spaceID,
		organizationID: organizationID,
		ownerID:        ownerID,
		keyName:        keyName,
		name:           name,
		spaceType:      spaceType,
	}
}

// EventType returns the event type identifier.
func (e SpaceCreated) EventType() string { return EventTypeSpaceCreated }

// SpaceID returns the created space's ID.
func (e SpaceCreated) SpaceID() SpaceID { return e.spaceID }

// OrganizationID returns the organization ID.
func (e SpaceCreated) OrganizationID() OrganizationID { return e.organizationID }

// OwnerID returns the owner's user ID.
func (e SpaceCreated) OwnerID() AppUserID { return e.ownerID }

// KeyName returns the space key name.
func (e SpaceCreated) KeyName() string { return e.keyName }

// Name returns the space display name.
func (e SpaceCreated) Name() string { return e.name }

// SpaceType returns the space type string.
func (e SpaceCreated) SpaceType() string { return e.spaceType }
