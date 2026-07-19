package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSyncRoute(t *testing.T) {
	t.Parallel()
	now := time.Now()
	id := uuid.Must(uuid.NewV7())
	ws := uuid.Must(uuid.NewV7())
	src := uuid.Must(uuid.NewV7())
	dest := uuid.Must(uuid.NewV7())

	route, err := NewSyncRoute(NewSyncRouteParams{
		ID:                 id,
		WorkspaceID:        ws,
		SourceConnectionID: src,
		Name:               "Nhanh → MISA invoice",
		Trigger:            SyncRouteTriggerWebhook,
		ResourceTypes:      []string{"Invoice", "invoice"},
		Activities: []ActivityGroup{{
			Mode: ActivityGroupSequential,
			Steps: []ActivityStep{{
				DestinationConnectionID: dest,
				Capability:              "invoice.create",
			}},
		}},
		Now: now,
	})
	if err != nil {
		t.Fatalf("NewSyncRoute: %v", err)
	}
	if route.Status() != SyncRouteStatusDraft {
		t.Fatalf("status = %s", route.Status())
	}
	if len(route.ResourceTypes()) != 1 || route.ResourceTypes()[0] != "invoice" {
		t.Fatalf("resource types = %v", route.ResourceTypes())
	}
	if !route.NeedsWebhookSecret() {
		t.Fatal("webhook trigger should need secret")
	}
}

func TestSyncRoute_EnableRequiresWebhookSecret(t *testing.T) {
	t.Parallel()
	route := mustNewRoute(t, SyncRouteTriggerWebhook, nil)
	if err := route.Enable(time.Now()); err != ErrWebhookSecretRequired {
		t.Fatalf("Enable = %v, want ErrWebhookSecretRequired", err)
	}
	if err := route.AttachWebhookSecretCiphertext([]byte("cipher"), time.Now()); err != nil {
		t.Fatalf("AttachWebhookSecretCiphertext: %v", err)
	}
	if err := route.Enable(time.Now()); err != nil {
		t.Fatalf("Enable after secret: %v", err)
	}
	if route.Status() != SyncRouteStatusEnabled {
		t.Fatalf("status = %s", route.Status())
	}
}

func TestSyncRoute_EnableRequiresSchedule(t *testing.T) {
	t.Parallel()
	route := mustNewRoute(t, SyncRouteTriggerSchedule, nil)
	if err := route.Enable(time.Now()); err != ErrInvalidSchedule {
		t.Fatalf("Enable = %v, want ErrInvalidSchedule", err)
	}
	_ = route.Update(UpdateSyncRouteParams{
		Schedule: map[string]any{"interval_seconds": 300},
		Now:      time.Now(),
	})
	if err := route.Enable(time.Now()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
}

func TestSyncRoute_UpdateBlockedWhenEnabled(t *testing.T) {
	t.Parallel()
	route := mustNewRoute(t, SyncRouteTriggerSchedule, map[string]any{"interval_seconds": 60})
	if err := route.Enable(time.Now()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	name := "other"
	if err := route.Update(UpdateSyncRouteParams{Name: &name, Now: time.Now()}); err != ErrSyncRouteNotEditable {
		t.Fatalf("Update = %v, want ErrSyncRouteNotEditable", err)
	}
	if err := route.Disable(time.Now()); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if err := route.Update(UpdateSyncRouteParams{Name: &name, Now: time.Now()}); err != nil {
		t.Fatalf("Update after disable: %v", err)
	}
}

func TestSyncRoute_InvalidActivities(t *testing.T) {
	t.Parallel()
	_, err := NewSyncRoute(NewSyncRouteParams{
		ID:                 uuid.Must(uuid.NewV7()),
		WorkspaceID:        uuid.Must(uuid.NewV7()),
		SourceConnectionID: uuid.Must(uuid.NewV7()),
		Name:               "bad",
		Trigger:            SyncRouteTriggerWebhook,
		ResourceTypes:      []string{"invoice"},
		Activities:         nil,
		Now:                time.Now(),
	})
	if err != ErrInvalidActivityGroup {
		t.Fatalf("got %v", err)
	}
}

func mustNewRoute(t *testing.T, trigger SyncRouteTrigger, schedule map[string]any) *SyncRoute {
	t.Helper()
	route, err := NewSyncRoute(NewSyncRouteParams{
		ID:                 uuid.Must(uuid.NewV7()),
		WorkspaceID:        uuid.Must(uuid.NewV7()),
		SourceConnectionID: uuid.Must(uuid.NewV7()),
		Name:               "route",
		Trigger:            trigger,
		ResourceTypes:      []string{"invoice"},
		Schedule:           schedule,
		Activities: []ActivityGroup{{
			Mode: ActivityGroupParallel,
			Steps: []ActivityStep{{
				DestinationConnectionID: uuid.Must(uuid.NewV7()),
				Capability:              "invoice.create",
			}},
		}},
		Now: time.Now(),
	})
	if err != nil {
		t.Fatalf("NewSyncRoute: %v", err)
	}
	return route
}
