package application

import (
	"context"

	"hublio/internal/events/domain"
)

// subscription pairs a match filter with a handler. match is either an exact event_name
// (e.g. "ExecutionSucceeded"), a Category (e.g. "runtime"), or "*" for every event.
type subscription struct {
	match   string
	handler domain.EventHandler
}

// Subscribe registers an in-process, at-least-once handler for events whose EventName or
// Category equals match, or every event when match is "*". Kept intentionally minimal: this
// is in-process delivery only (no durable subscription state, no consumer groups) — the
// Event Platform is not a message broker.
func (s *Services) Subscribe(match string, handler domain.EventHandler) {
	if handler == nil {
		return
	}
	s.subscriptions = append(s.subscriptions, subscription{match: match, handler: handler})
}

// notifySubscribers calls every matching handler after the event has already been
// persisted. Handler errors are never surfaced to the publisher (at-least-once, best
// effort): a subscriber that needs stronger guarantees must implement its own retry/DLQ.
func (s *Services) notifySubscribers(ctx context.Context, event *domain.PlatformEvent) {
	for _, sub := range s.subscriptions {
		if !subscriptionMatches(sub.match, event) {
			continue
		}
		if err := sub.handler(ctx, event); err != nil && s.OnSubscriberError != nil {
			s.OnSubscriberError(event, err)
		}
	}
}

func subscriptionMatches(match string, event *domain.PlatformEvent) bool {
	if match == "*" {
		return true
	}
	if match == event.EventName() {
		return true
	}
	return match == string(event.Category())
}
