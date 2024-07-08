package realtimeservice

import (
	"bytes"
	"fmt"
)

type WSContentEvent int

const (
	WSContentEventInvalid WSContentEvent = iota
	WSContentEventAll
	WSContentEventCreate
	WSContentEventUpdate
	WSContentEventDelete
	endWSContentEvents
)

var (
	wsContentEventToStrings = map[WSContentEvent]string{
		WSContentEventInvalid: "invalid",
		WSContentEventAll:     "*",
		WSContentEventCreate:  "create",
		WSContentEventUpdate:  "update",
		WSContentEventDelete:  "delete",
	}

	stringToWSContentEvents = map[string]WSContentEvent{
		"invalid": WSContentEventInvalid,
		"*":       WSContentEventAll,
		"create":  WSContentEventCreate,
		"update":  WSContentEventUpdate,
		"delete":  WSContentEventDelete,
	}
)

func WSContentEventFromString(s string) WSContentEvent {
	if t, ok := stringToWSContentEvents[s]; ok {
		return t
	}
	return WSContentEventInvalid
}

// String returns the string value of the type
func (t WSContentEvent) String() string {
	if t < endWSContentEvents {
		return wsContentEventToStrings[t]
	}
	return wsContentEventToStrings[WSContentEventInvalid]
}

// Valid reports if the given type if known type.
func (t WSContentEvent) Valid() bool {
	return t > WSContentEventInvalid && t < endWSContentEvents
}

// MarshalJSON marshals the enum as a quoted json string
func (t WSContentEvent) MarshalJSON() ([]byte, error) {
	eventStr, ok := wsContentEventToStrings[t]
	if !ok {
		return nil, fmt.Errorf("unknown WSContentEvent value: %d", t)
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(eventStr)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}
