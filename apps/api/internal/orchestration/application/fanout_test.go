package application

import (
	"testing"
	"time"

	"hublio/internal/orchestration/domain"

	"github.com/google/uuid"
)

func TestExecutionTarget_FanOutOverrides(t *testing.T) {
	t.Parallel()
	intentConn := uuid.Must(uuid.NewV7())
	fanConn := uuid.Must(uuid.NewV7())
	intent := mustIntent(t, intentConn, "invoice.create")
	exec := mustExecution(t, intent.ID())
	exec.MergeContext(map[string]any{
		CtxFanoutConnectionID: fanConn.String(),
		CtxFanoutCapability:   "invoice.update_status",
	})
	gotConn, gotCap := executionTarget(intent, exec)
	if gotConn != fanConn {
		t.Fatalf("connection: got %s want %s", gotConn, fanConn)
	}
	if gotCap != "invoice.update_status" {
		t.Fatalf("capability: got %q", gotCap)
	}
}

func TestExecutionTarget_FallsBackToIntent(t *testing.T) {
	t.Parallel()
	intentConn := uuid.Must(uuid.NewV7())
	intent := mustIntent(t, intentConn, "invoice.create")
	exec := mustExecution(t, intent.ID())
	gotConn, gotCap := executionTarget(intent, exec)
	if gotConn != intentConn || gotCap != "invoice.create" {
		t.Fatalf("got %s %q", gotConn, gotCap)
	}
}

func TestStringSliceFromAny(t *testing.T) {
	t.Parallel()
	a, ok := stringSliceFromAny([]string{"a", "b"})
	if !ok || len(a) != 2 || a[0] != "a" {
		t.Fatalf("[]string: %#v", a)
	}
	b, ok := stringSliceFromAny([]any{"x", "y"})
	if !ok || len(b) != 2 || b[1] != "y" {
		t.Fatalf("[]any: %#v", b)
	}
	if _, ok := stringSliceFromAny("nope"); ok {
		t.Fatal("expected false")
	}
}

func TestIntFromAny(t *testing.T) {
	t.Parallel()
	if intFromAny(0) != 0 || intFromAny(float64(2)) != 2 || intFromAny("x") != -1 {
		t.Fatal("intFromAny mismatch")
	}
}

func mustIntent(t *testing.T, connectionID uuid.UUID, capability string) *domain.Intent {
	t.Helper()
	intent, err := domain.NewIntent(
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		connectionID,
		capability,
		map[string]any{"id": "1"},
		"corr",
		"",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("NewIntent: %v", err)
	}
	return intent
}

func mustExecution(t *testing.T, intentID uuid.UUID) *domain.Execution {
	t.Helper()
	stepIDs := make([]uuid.UUID, len(domain.DefaultStepTypes()))
	for i := range stepIDs {
		stepIDs[i] = uuid.Must(uuid.NewV7())
	}
	exec, err := domain.NewExecution(uuid.Must(uuid.NewV7()), intentID, stepIDs, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewExecution: %v", err)
	}
	return exec
}
