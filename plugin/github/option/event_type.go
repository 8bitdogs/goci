package option

import "strings"

var _ Option = (*eventTypeOption)(nil)

type EventType string

const (
	EventTypePush     = "push"
	EventTypeWorkflow = "workflow_job"
)

type eventTypeOption struct {
	eventType string
}

func (o *eventTypeOption) Apply(opts *Options) {
	opts.eventType = o.eventType
}

func WithEventType(eventType string) Option {
	lwET := strings.ToLower(eventType)
	switch lwET {
	case EventTypePush, EventTypeWorkflow:
		return &eventTypeOption{eventType: lwET}
	default:
		panic("invalid event type: " + eventType)
	}
}
