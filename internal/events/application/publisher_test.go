package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"hublio/internal/events/domain"
	"hublio/internal/platform/metrics"

	"github.com/google/uuid"
)

type fakeEventRepository struct {
	saved   []*domain.PlatformEvent
	saveErr error
}

func (f *fakeEventRepository) Save(ctx context.Context, event *domain.PlatformEvent) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, event)
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestPublish_PersistsBeforeNotifyingSubscribers(t *testing.T) {
	repo := &fakeEventRepository{}
	counters := metrics.New()
	svc := &Services{
		Events:  repo,
		Metrics: counters,
		Clock:   fixedClock{now: time.Now()},
	}

	var subscriberSawSavedEvent bool
	svc.Subscribe("*", func(ctx context.Context, event *domain.PlatformEvent) error {
		// By the time a subscriber runs, Save must have already happened (Publish
		// persists first, then notifies — see AGENTS.md/checklist F1).
		subscriberSawSavedEvent = len(repo.saved) == 1 && repo.saved[0] == event
		return nil
	})

	aggregateID := uuid.Must(uuid.NewV7())
	err := svc.Publish(context.Background(), PublishInput{
		AggregateType: string(domain.AggregateTypeExecution),
		AggregateID:   aggregateID,
		Category:      string(domain.CategoryRuntime),
		EventName:     "ExecutionSucceeded",
		PublishedBy:   "orchestration",
	})
	if err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved event, got %d", len(repo.saved))
	}
	if repo.saved[0].AggregateID() != aggregateID {
		t.Fatalf("saved AggregateID = %v, want %v", repo.saved[0].AggregateID(), aggregateID)
	}
	if !subscriberSawSavedEvent {
		t.Fatal("subscriber ran before the event was persisted")
	}
	if counters.Snapshot().EventsPublished != 1 {
		t.Fatalf("events_published_total = %d, want 1", counters.Snapshot().EventsPublished)
	}
}

func TestPublish_SubscriberErrorDoesNotFailPublish(t *testing.T) {
	repo := &fakeEventRepository{}
	var reportedErr error
	svc := &Services{
		Events: repo,
		OnSubscriberError: func(event *domain.PlatformEvent, err error) {
			reportedErr = err
		},
	}
	svc.Subscribe("*", func(ctx context.Context, event *domain.PlatformEvent) error {
		return errors.New("boom")
	})

	err := svc.Publish(context.Background(), PublishInput{
		AggregateType: string(domain.AggregateTypeConnection),
		AggregateID:   uuid.Must(uuid.NewV7()),
		Category:      string(domain.CategorySystem),
		EventName:     "ConnectionEnabled",
	})
	if err != nil {
		t.Fatalf("Publish() must not fail on subscriber error, got: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected event to still be persisted, got %d saved", len(repo.saved))
	}
	if reportedErr == nil || reportedErr.Error() != "boom" {
		t.Fatalf("OnSubscriberError not invoked with subscriber error, got: %v", reportedErr)
	}
}

func TestPublish_RepositoryFailureIsSurfaced(t *testing.T) {
	repo := &fakeEventRepository{saveErr: errors.New("db down")}
	svc := &Services{Events: repo}

	err := svc.Publish(context.Background(), PublishInput{
		AggregateType: string(domain.AggregateTypeIntent),
		AggregateID:   uuid.Must(uuid.NewV7()),
		Category:      string(domain.CategoryRuntime),
		EventName:     "IntentAccepted",
	})
	if err == nil {
		t.Fatal("expected Publish() to surface the repository error")
	}
}

func TestPublish_InvalidInputRejected(t *testing.T) {
	repo := &fakeEventRepository{}
	svc := &Services{Events: repo}

	err := svc.Publish(context.Background(), PublishInput{
		AggregateType: string(domain.AggregateTypeExecution),
		AggregateID:   uuid.Nil, // invalid: NewPlatformEvent requires a non-nil aggregate id
		Category:      string(domain.CategoryRuntime),
		EventName:     "ExecutionSucceeded",
	})
	if err == nil {
		t.Fatal("expected Publish() to reject an invalid PublishInput")
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected nothing persisted on validation failure, got %d", len(repo.saved))
	}
}

func TestSubscribe_MatchesByEventNameAndCategory(t *testing.T) {
	repo := &fakeEventRepository{}
	svc := &Services{Events: repo}

	var byName, byCategory, byWildcard int
	svc.Subscribe("ExecutionSucceeded", func(ctx context.Context, event *domain.PlatformEvent) error {
		byName++
		return nil
	})
	svc.Subscribe(string(domain.CategoryRuntime), func(ctx context.Context, event *domain.PlatformEvent) error {
		byCategory++
		return nil
	})
	svc.Subscribe("*", func(ctx context.Context, event *domain.PlatformEvent) error {
		byWildcard++
		return nil
	})

	if err := svc.Publish(context.Background(), PublishInput{
		AggregateType: string(domain.AggregateTypeExecution),
		AggregateID:   uuid.Must(uuid.NewV7()),
		Category:      string(domain.CategoryRuntime),
		EventName:     "ExecutionSucceeded",
	}); err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}

	if byName != 1 || byCategory != 1 || byWildcard != 1 {
		t.Fatalf("byName=%d byCategory=%d byWildcard=%d, want 1/1/1", byName, byCategory, byWildcard)
	}
}
