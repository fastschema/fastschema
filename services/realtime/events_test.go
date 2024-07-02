package realtimeservice_test

import (
	"testing"

	realtimeservice "github.com/fastschema/fastschema/services/realtime"
	"github.com/stretchr/testify/assert"
)

func TestWSContentEventFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected realtimeservice.WSContentEvent
	}{
		{input: "invalid", expected: realtimeservice.WSContentEventInvalid},
		{input: "*", expected: realtimeservice.WSContentEventAll},
		{input: "create", expected: realtimeservice.WSContentEventCreate},
		{input: "update", expected: realtimeservice.WSContentEventUpdate},
		{input: "delete", expected: realtimeservice.WSContentEventDelete},
		{input: "unknown", expected: realtimeservice.WSContentEventInvalid},
	}

	for _, test := range tests {
		result := realtimeservice.WSContentEventFromString(test.input)
		assert.Equal(t, test.expected, result)
	}
}
func TestWSContentEventString(t *testing.T) {
	tests := []struct {
		event    realtimeservice.WSContentEvent
		expected string
	}{
		{event: realtimeservice.WSContentEventInvalid, expected: "invalid"},
		{event: realtimeservice.WSContentEventAll, expected: "*"},
		{event: realtimeservice.WSContentEventCreate, expected: "create"},
		{event: realtimeservice.WSContentEventUpdate, expected: "update"},
		{event: realtimeservice.WSContentEventDelete, expected: "delete"},
		{event: realtimeservice.WSContentEvent(42), expected: "invalid"},
	}

	for _, test := range tests {
		result := test.event.String()
		assert.Equal(t, test.expected, result)
	}
}

func TestWSContentEventValid(t *testing.T) {
	tests := []struct {
		event    realtimeservice.WSContentEvent
		expected bool
	}{
		{event: realtimeservice.WSContentEventInvalid, expected: false},
		{event: realtimeservice.WSContentEventAll, expected: true},
		{event: realtimeservice.WSContentEventCreate, expected: true},
		{event: realtimeservice.WSContentEventUpdate, expected: true},
		{event: realtimeservice.WSContentEventDelete, expected: true},
		{event: realtimeservice.WSContentEvent(42), expected: false},
	}
	for _, test := range tests {
		result := test.event.Valid()
		assert.Equal(t, test.expected, result)
	}
}

func TestWSContentEventMarshalJSON(t *testing.T) {
	tests := []struct {
		event    realtimeservice.WSContentEvent
		expected []byte
	}{
		{event: realtimeservice.WSContentEventInvalid, expected: []byte(`"invalid"`)},
		{event: realtimeservice.WSContentEventAll, expected: []byte(`"*"`)},
		{event: realtimeservice.WSContentEventCreate, expected: []byte(`"create"`)},
		{event: realtimeservice.WSContentEventUpdate, expected: []byte(`"update"`)},
		{event: realtimeservice.WSContentEventDelete, expected: []byte(`"delete"`)},
	}
	for _, test := range tests {
		result, err := test.event.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, result)
	}

	_, err := realtimeservice.WSContentEvent(42).MarshalJSON()
	assert.Error(t, err)
}
