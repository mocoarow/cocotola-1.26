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
	appUserID      AppUserID
	organizationID OrganizationID
	loginID        string
	occurredAt     time.Time
}

// NewAppUserCreated returns a new AppUserCreated event.
func NewAppUserCreated(appUserID AppUserID, organizationID OrganizationID, loginID string, occurredAt time.Time) AppUserCreated {
	return AppUserCreated{
		appUserID:      appUserID,
		organizationID: organizationID,
		loginID:        loginID,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e AppUserCreated) EventType() string { return EventTypeAppUserCreated }

// OccurredAt returns the time the event occurred.
func (e AppUserCreated) OccurredAt() time.Time { return e.occurredAt }

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
	groupID        GroupID
	organizationID OrganizationID
	name           string
	occurredAt     time.Time
}

// NewGroupCreated returns a new GroupCreated event.
func NewGroupCreated(groupID GroupID, organizationID OrganizationID, name string, occurredAt time.Time) GroupCreated {
	return GroupCreated{
		groupID:        groupID,
		organizationID: organizationID,
		name:           name,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e GroupCreated) EventType() string { return EventTypeGroupCreated }

// OccurredAt returns the time the event occurred.
func (e GroupCreated) OccurredAt() time.Time { return e.occurredAt }

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
	spaceID        SpaceID
	organizationID OrganizationID
	ownerID        AppUserID
	keyName        string
	name           string
	spaceType      string
	occurredAt     time.Time
}

// NewSpaceCreated returns a new SpaceCreated event.
func NewSpaceCreated(spaceID SpaceID, organizationID OrganizationID, ownerID AppUserID, keyName string, name string, spaceType string, occurredAt time.Time) SpaceCreated {
	return SpaceCreated{
		spaceID:        spaceID,
		organizationID: organizationID,
		ownerID:        ownerID,
		keyName:        keyName,
		name:           name,
		spaceType:      spaceType,
		occurredAt:     occurredAt,
	}
}

// EventType returns the event type identifier.
func (e SpaceCreated) EventType() string { return EventTypeSpaceCreated }

// OccurredAt returns the time the event occurred.
func (e SpaceCreated) OccurredAt() time.Time { return e.occurredAt }

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
