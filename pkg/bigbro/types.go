package bigbro

import (
	"maps"
	"strconv"
)

const (
	analyticsEventPath = "/analytics/event"

	EventTypeClick  EventType = "click"
	EventTypeAction EventType = "action"
	EventTypeNavgo  EventType = "navgo"
	EventTypeView   EventType = "view"
)

type (
	EventType   string
	EventFields map[string]any

	event struct {
		Event       string      `json:"event"`
		Type        EventType   `json:"type"`
		UserID      string      `json:"user_id"`
		EventFields EventFields `json:"json,omitempty"` //nolint:tagliatelle
	}

	eventResponse struct {
		OK bool `json:"ok"`
	}
)

func eventBuilder(eventName string, eventType EventType, userID int64, eventFields, configFields EventFields) event {
	baseFields := make(EventFields, len(eventFields)+len(configFields))
	maps.Copy(baseFields, eventFields)
	maps.Copy(baseFields, configFields)

	return event{
		Event:       eventName,
		Type:        eventType,
		UserID:      strconv.FormatInt(userID, 10),
		EventFields: baseFields,
	}
}
