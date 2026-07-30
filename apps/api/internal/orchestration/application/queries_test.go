package application_test

import (
	"context"
	"testing"
	"time"

	"hublio/internal/orchestration/application"
	"hublio/internal/orchestration/domain"

	"github.com/google/uuid"
)

type memIntents struct {
	byID map[uuid.UUID]*domain.Intent
}

func newMemIntents() *memIntents {
	return &memIntents{byID: map[uuid.UUID]*domain.Intent{}}
}

func (m *memIntents) Save(ctx context.Context, intent *domain.Intent) error {
	_ = ctx
	m.byID[intent.ID()] = intent
	return nil
}

func (m *memIntents) Update(ctx context.Context, intent *domain.Intent) error {
	return m.Save(ctx, intent)
}

func (m *memIntents) FindByID(ctx context.Context, id uuid.UUID) (*domain.Intent, error) {
	_ = ctx
	intent, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return intent, nil
}

func (m *memIntents) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, status *string, limit int32) ([]*domain.Intent, error) {
	_ = ctx
	out := make([]*domain.Intent, 0)
	for _, intent := range m.byID {
		if intent.WorkspaceID() != workspaceID {
			continue
		}
		if status != nil && string(intent.Status()) != *status {
			continue
		}
		out = append(out, intent)
	}
	if int(limit) < len(out) {
		out = out[:limit]
	}
	return out, nil
}

type memExecutions struct {
	byID     map[uuid.UUID]*domain.Execution
	intents  *memIntents
	listRows []*domain.Execution
}

func newMemExecutions(intents *memIntents) *memExecutions {
	return &memExecutions{byID: map[uuid.UUID]*domain.Execution{}, intents: intents}
}

func (m *memExecutions) Save(ctx context.Context, execution *domain.Execution) error {
	_ = ctx
	m.byID[execution.ID()] = execution
	return nil
}

func (m *memExecutions) Update(ctx context.Context, execution *domain.Execution) error {
	return m.Save(ctx, execution)
}

func (m *memExecutions) FindByID(ctx context.Context, id uuid.UUID) (*domain.Execution, error) {
	_ = ctx
	exec, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return exec, nil
}

func (m *memExecutions) FindByIntentID(ctx context.Context, intentID uuid.UUID) (*domain.Execution, error) {
	_ = ctx
	for _, exec := range m.byID {
		if exec.IntentID() == intentID {
			return exec, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *memExecutions) ListByIntentID(ctx context.Context, intentID uuid.UUID) ([]*domain.Execution, error) {
	_ = ctx
	out := make([]*domain.Execution, 0)
	for _, exec := range m.byID {
		if exec.IntentID() == intentID {
			out = append(out, exec)
		}
	}
	return out, nil
}

func (m *memExecutions) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, status *string, limit int32) ([]*domain.Execution, error) {
	_ = ctx
	source := m.listRows
	if source == nil {
		source = make([]*domain.Execution, 0, len(m.byID))
		for _, exec := range m.byID {
			source = append(source, exec)
		}
	}
	out := make([]*domain.Execution, 0)
	for _, exec := range source {
		intent, err := m.intents.FindByID(ctx, exec.IntentID())
		if err != nil || intent.WorkspaceID() != workspaceID {
			continue
		}
		if status != nil && string(exec.Status()) != *status {
			continue
		}
		out = append(out, exec)
	}
	if int(limit) < len(out) {
		out = out[:limit]
	}
	return out, nil
}

func mustListIntent(t *testing.T, workspaceID uuid.UUID, status domain.IntentStatus) *domain.Intent {
	t.Helper()
	intent, err := domain.NewIntent(
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		workspaceID,
		uuid.Must(uuid.NewV7()),
		"echo",
		map[string]any{"msg": "hi"},
		"corr",
		uuid.Must(uuid.NewV7()).String(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("NewIntent: %v", err)
	}
	switch status {
	case domain.IntentStatusAccepted:
		intent.Accept(time.Now().UTC())
	case domain.IntentStatusRejected:
		intent.Reject("test", time.Now().UTC())
	case domain.IntentStatusExpired:
		intent.Expire(time.Now().UTC())
	}
	return intent
}

func mustListExecution(t *testing.T, intentID uuid.UUID) *domain.Execution {
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

func TestListIntents(t *testing.T) {
	t.Parallel()
	ws := uuid.Must(uuid.NewV7())
	other := uuid.Must(uuid.NewV7())
	intents := newMemIntents()
	a := mustListIntent(t, ws, domain.IntentStatusAccepted)
	b := mustListIntent(t, ws, domain.IntentStatusRejected)
	c := mustListIntent(t, other, domain.IntentStatusAccepted)
	_ = intents.Save(context.Background(), a)
	_ = intents.Save(context.Background(), b)
	_ = intents.Save(context.Background(), c)

	svc := &application.Services{Intents: intents}

	t.Run("clamps limit default", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListIntents(context.Background(), ws, nil, 0)
		if err != nil {
			t.Fatalf("ListIntents: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len=%d want 2", len(got))
		}
	})

	t.Run("clamps limit max", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListIntents(context.Background(), ws, nil, 999)
		if err != nil {
			t.Fatalf("ListIntents: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len=%d want 2", len(got))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		t.Parallel()
		status := string(domain.IntentStatusRejected)
		got, err := svc.ListIntents(context.Background(), ws, &status, 50)
		if err != nil {
			t.Fatalf("ListIntents: %v", err)
		}
		if len(got) != 1 || got[0].ID() != b.ID() {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("empty workspace", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListIntents(context.Background(), uuid.Must(uuid.NewV7()), nil, 50)
		if err != nil {
			t.Fatalf("ListIntents: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("len=%d want 0", len(got))
		}
	})
}

func TestListExecutions(t *testing.T) {
	t.Parallel()
	ws := uuid.Must(uuid.NewV7())
	intents := newMemIntents()
	execs := newMemExecutions(intents)

	intentOK := mustListIntent(t, ws, domain.IntentStatusAccepted)
	intentOther := mustListIntent(t, uuid.Must(uuid.NewV7()), domain.IntentStatusAccepted)
	_ = intents.Save(context.Background(), intentOK)
	_ = intents.Save(context.Background(), intentOther)

	e1 := mustListExecution(t, intentOK.ID())
	e2 := mustListExecution(t, intentOK.ID())
	eOther := mustListExecution(t, intentOther.ID())
	_ = e1.Queue(time.Now().UTC())
	_ = execs.Save(context.Background(), e1)
	_ = execs.Save(context.Background(), e2)
	_ = execs.Save(context.Background(), eOther)

	svc := &application.Services{Intents: intents, Executions: execs}

	t.Run("scopes by workspace", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListExecutions(context.Background(), ws, nil, 50)
		if err != nil {
			t.Fatalf("ListExecutions: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len=%d want 2", len(got))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		t.Parallel()
		status := string(domain.ExecutionStatusQueued)
		got, err := svc.ListExecutions(context.Background(), ws, &status, 50)
		if err != nil {
			t.Fatalf("ListExecutions: %v", err)
		}
		if len(got) != 1 || got[0].ID() != e1.ID() {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("empty workspace", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListExecutions(context.Background(), uuid.Must(uuid.NewV7()), nil, 10)
		if err != nil {
			t.Fatalf("ListExecutions: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("len=%d want 0", len(got))
		}
	})
}

func TestClampListLimit(t *testing.T) {
	t.Parallel()
	ws := uuid.Must(uuid.NewV7())
	intents := newMemIntents()
	for i := 0; i < 5; i++ {
		_ = intents.Save(context.Background(), mustListIntent(t, ws, domain.IntentStatusAccepted))
	}
	svc := &application.Services{Intents: intents}

	got, err := svc.ListIntents(context.Background(), ws, nil, 2)
	if err != nil {
		t.Fatalf("ListIntents: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
}
