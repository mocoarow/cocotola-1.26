package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var (
	fixtureEventUserID  = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
	fixtureEventGroupID = domain.MustParseGroupID("00000000-0000-7000-8000-000000000030")
	fixtureEventSpaceID = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000040")
)

type eventCase struct {
	name       string
	event      domain.Event
	occurredAt time.Time
	eventType  string
}

func eventCases() []eventCase {
	appUserCreatedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	groupCreatedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	spaceCreatedAt := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

	return []eventCase{
		{
			name:       "AppUserCreated",
			event:      domain.NewAppUserCreated(fixtureEventUserID, fixtureOrgID, "user1", appUserCreatedAt),
			occurredAt: appUserCreatedAt,
			eventType:  domain.EventTypeAppUserCreated,
		},
		{
			name:       "GroupCreated",
			event:      domain.NewGroupCreated(fixtureEventGroupID, fixtureOrgID, "group1", groupCreatedAt),
			occurredAt: groupCreatedAt,
			eventType:  domain.EventTypeGroupCreated,
		},
		{
			name:       "SpaceCreated",
			event:      domain.NewSpaceCreated(fixtureEventSpaceID, fixtureOrgID, fixtureEventUserID, "key1", "space1", "private", spaceCreatedAt),
			occurredAt: spaceCreatedAt,
			eventType:  domain.EventTypeSpaceCreated,
		},
	}
}

func Test_Event_OccurredAt_shouldReturnTimeGivenToConstructor_whenEventIsCreated(t *testing.T) {
	t.Parallel()

	for _, tt := range eventCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			event := tt.event

			// when
			occurredAt := event.OccurredAt()

			// then
			assert.Equal(t, tt.occurredAt, occurredAt)
		})
	}
}

func Test_Event_EventType_shouldReturnTypeConstant_whenEventIsCreated(t *testing.T) {
	t.Parallel()

	for _, tt := range eventCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			event := tt.event

			// when
			eventType := event.EventType()

			// then
			assert.Equal(t, tt.eventType, eventType)
		})
	}
}
