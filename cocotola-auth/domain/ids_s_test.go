package domain_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

const (
	validIDString   = "0192d3f0-7b3a-7c4e-8f1a-2b3c4d5e6f70"
	invalidIDString = "not-a-uuid"
)

type idObject interface {
	String() string
	IsZero() bool
}

type idValuer interface {
	idObject
	driver.Valuer
}

type idScanner[T any] interface {
	*T
	sql.Scanner
}

type idEnvelope[T any] struct {
	ID T `json:"id"`
}

type idSpec struct {
	name       string
	parse      func(s string) (idObject, error)
	scan       func(src any) (idObject, error)
	value      func(s string) (driver.Value, error)
	encodeJSON func(s string) ([]byte, error)
	decodeJSON func(data []byte) (idObject, error)
}

func newIDSpec[T idValuer, P idScanner[T]](name string, parse func(string) (T, error)) idSpec {
	scan := func(src any) (T, error) {
		var v T
		if err := P(&v).Scan(src); err != nil {
			return v, fmt.Errorf("scan: %w", err)
		}
		return v, nil
	}
	decodeJSON := func(data []byte) (T, error) {
		var v idEnvelope[T]
		if err := json.Unmarshal(data, &v); err != nil {
			return v.ID, fmt.Errorf("unmarshal: %w", err)
		}
		return v.ID, nil
	}

	return idSpec{
		name:       name,
		parse:      asIDFunc(parse),
		scan:       asIDFunc(scan),
		decodeJSON: asIDFunc(decodeJSON),
		value: func(s string) (driver.Value, error) {
			src, err := parse(s)
			if err != nil {
				return nil, fmt.Errorf("parse: %w", err)
			}
			v, err := src.Value()
			if err != nil {
				return nil, fmt.Errorf("value: %w", err)
			}
			return v, nil
		},
		encodeJSON: func(s string) ([]byte, error) {
			src, err := parse(s)
			if err != nil {
				return nil, fmt.Errorf("parse: %w", err)
			}
			data, err := json.Marshal(idEnvelope[T]{ID: src})
			if err != nil {
				return nil, fmt.Errorf("marshal: %w", err)
			}
			return data, nil
		},
	}
}

func asIDFunc[A any, T idObject](f func(A) (T, error)) func(A) (idObject, error) {
	return func(a A) (idObject, error) {
		v, err := f(a)
		return v, err
	}
}

func idSpecs() []idSpec {
	return []idSpec{
		newIDSpec("OrganizationID", domain.ParseOrganizationID),
		newIDSpec("AppUserID", domain.ParseAppUserID),
		newIDSpec("GroupID", domain.ParseGroupID),
		newIDSpec("SpaceID", domain.ParseSpaceID),
		newIDSpec("AppUserProviderID", domain.ParseAppUserProviderID),
	}
}

func Test_ParseID_shouldReturnID_whenUUIDIsValid(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			got, err := tt.parse(validIDString)

			// then
			require.NoError(t, err)
			assert.Equal(t, validIDString, got.String())
		})
	}
}

func Test_ParseID_shouldReturnInvalidArgument_whenEmpty(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			_, err := tt.parse("")

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_ParseID_shouldReturnInvalidArgument_whenUUIDIsMalformed(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			_, err := tt.parse(invalidIDString)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_ID_Scan_shouldReturnZeroID_whenSrcIsNil(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			got, err := tt.scan(nil)

			// then
			require.NoError(t, err)
			assert.True(t, got.IsZero())
		})
	}
}

func Test_ID_Scan_shouldReturnID_whenSrcIsValidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  any
	}{
		{name: "string", src: validIDString},
		{name: "bytes", src: []byte(validIDString)},
	}
	for _, spec := range idSpecs() {
		for _, tt := range tests {
			t.Run(spec.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				// given / when
				got, err := spec.scan(tt.src)

				// then
				require.NoError(t, err)
				assert.Equal(t, validIDString, got.String())
			})
		}
	}
}

func Test_ID_Scan_shouldReturnInvalidArgument_whenSrcIsMalformedUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  any
	}{
		{name: "string", src: invalidIDString},
		{name: "bytes", src: []byte(invalidIDString)},
	}
	for _, spec := range idSpecs() {
		for _, tt := range tests {
			t.Run(spec.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				// given / when
				_, err := spec.scan(tt.src)

				// then
				require.ErrorIs(t, err, domain.ErrInvalidArgument)
			})
		}
	}
}

func Test_ID_Scan_shouldReturnInvalidArgument_whenSrcTypeIsUnsupported(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			_, err := tt.scan(int64(1))

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_ID_Value_shouldReturnUUIDString(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			got, err := tt.value(validIDString)

			// then
			require.NoError(t, err)
			assert.Equal(t, driver.Value(validIDString), got)
		})
	}
}

func Test_ID_ValueThenScan_shouldPreserveID(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			v, err := tt.value(validIDString)
			require.NoError(t, err)

			// when
			got, err := tt.scan(v)

			// then
			require.NoError(t, err)
			assert.Equal(t, validIDString, got.String())
		})
	}
}

func Test_ID_MarshalJSON_shouldEncodeAsUUIDString(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			got, err := tt.encodeJSON(validIDString)

			// then
			require.NoError(t, err)
			assert.JSONEq(t, `{"id":"`+validIDString+`"}`, string(got))
		})
	}
}

func Test_ID_MarshalThenUnmarshalJSON_shouldPreserveID(t *testing.T) {
	t.Parallel()

	for _, tt := range idSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			data, err := tt.encodeJSON(validIDString)
			require.NoError(t, err)

			// when
			got, err := tt.decodeJSON(data)

			// then
			require.NoError(t, err)
			assert.Equal(t, validIDString, got.String())
		})
	}
}

func Test_ID_UnmarshalJSON_shouldReturnInvalidArgument_whenStringIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
	}{
		{name: "empty", data: `{"id":""}`},
		{name: "malformed", data: `{"id":"` + invalidIDString + `"}`},
	}
	for _, spec := range idSpecs() {
		for _, tt := range tests {
			t.Run(spec.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				// given / when
				_, err := spec.decodeJSON([]byte(tt.data))

				// then
				require.ErrorIs(t, err, domain.ErrInvalidArgument)
			})
		}
	}
}

type fromUUIDSpec struct {
	name     string
	fromUUID func(uuid.UUID) (idObject, error)
}

func fromUUIDSpecs() []fromUUIDSpec {
	return []fromUUIDSpec{
		{name: "OrganizationID", fromUUID: asIDFunc(domain.NewOrganizationIDFromUUID)},
		{name: "AppUserID", fromUUID: asIDFunc(domain.NewAppUserIDFromUUID)},
	}
}

func Test_NewIDFromUUID_shouldReturnID_whenUUIDIsNonZero(t *testing.T) {
	t.Parallel()

	for _, tt := range fromUUIDSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			u := uuid.MustParse(validIDString)

			// when
			got, err := tt.fromUUID(u)

			// then
			require.NoError(t, err)
			assert.Equal(t, validIDString, got.String())
		})
	}
}

func Test_NewIDFromUUID_shouldReturnInvalidArgument_whenUUIDIsZero(t *testing.T) {
	t.Parallel()

	for _, tt := range fromUUIDSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given / when
			_, err := tt.fromUUID(uuid.Nil)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
