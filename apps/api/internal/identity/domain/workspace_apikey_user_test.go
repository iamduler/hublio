package domain_test

import (
	"testing"
	"time"

	"hublio/internal/identity/domain"

	"github.com/google/uuid"
)

func TestWorkspaceEnableDisable(t *testing.T) {
	now := time.Now().UTC()
	orgID := uuid.MustParse("01900000-0000-7000-8000-000000000010")
	wsID := uuid.MustParse("01900000-0000-7000-8000-000000000011")

	ws, err := domain.NewWorkspace(wsID, orgID, "prod", "production", now)
	if err != nil {
		t.Fatal(err)
	}
	if !ws.CanExecuteIntents() {
		t.Fatal("expected active workspace to execute intents")
	}

	if err := ws.Disable(now); err != nil {
		t.Fatal(err)
	}
	if ws.CanExecuteIntents() {
		t.Fatal("disabled workspace must not execute intents")
	}

	if err := ws.Enable(now); err != nil {
		t.Fatal(err)
	}
	if !ws.CanExecuteIntents() {
		t.Fatal("enabled workspace should execute intents")
	}
}

func TestAPIKeyRotateAndDisable(t *testing.T) {
	now := time.Now().UTC()
	key, err := domain.NewAPIKey(
		uuid.MustParse("01900000-0000-7000-8000-000000000021"),
		uuid.MustParse("01900000-0000-7000-8000-000000000011"),
		"ci",
		"hash1",
		"abcd1234",
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := key.IsUsable(now); err != nil {
		t.Fatal(err)
	}

	if err := key.Rotate("hash2", "efgh5678", now); err != nil {
		t.Fatal(err)
	}
	if key.Prefix() != "efgh5678" || key.KeyHash() != "hash2" {
		t.Fatal("rotate did not update credentials")
	}

	events := key.PullEvents()
	if len(events) < 2 {
		t.Fatalf("expected create+rotate events, got %d", len(events))
	}

	if err := key.Disable(now); err != nil {
		t.Fatal(err)
	}
	if err := key.IsUsable(now); err != domain.ErrAPIKeyDisabled {
		t.Fatalf("got %v", err)
	}
}

func TestUserCanLogin(t *testing.T) {
	now := time.Now().UTC()
	u, err := domain.NewUser(
		uuid.MustParse("01900000-0000-7000-8000-000000000031"),
		uuid.MustParse("01900000-0000-7000-8000-000000000001"),
		"Owner@Example.com",
		"Owner",
		"hash",
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if u.Email() != "owner@example.com" {
		t.Fatalf("email not normalized: %s", u.Email())
	}
	if !u.CanLogin() {
		t.Fatal("expected can login")
	}
	if err := u.Suspend(now); err != nil {
		t.Fatal(err)
	}
	if u.CanLogin() {
		t.Fatal("suspended cannot login")
	}
}
