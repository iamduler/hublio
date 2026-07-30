package application_test

import (
	"context"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

type memMemberships struct {
	byKey   map[string]*domain.Membership
	members []*domain.WorkspaceMember
}

func newMemMemberships() *memMemberships {
	return &memMemberships{byKey: map[string]*domain.Membership{}}
}

func membershipKey(ws, user uuid.UUID) string {
	return ws.String() + ":" + user.String()
}

func (m *memMemberships) Save(ctx context.Context, membership *domain.Membership) error {
	_ = ctx
	m.byKey[membershipKey(membership.WorkspaceID(), membership.UserID())] = membership
	return nil
}

func (m *memMemberships) Find(ctx context.Context, workspaceID, userID uuid.UUID) (*domain.Membership, error) {
	_ = ctx
	mem, ok := m.byKey[membershipKey(workspaceID, userID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return mem, nil
}

func (m *memMemberships) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Membership, error) {
	_ = ctx
	out := make([]*domain.Membership, 0)
	for _, mem := range m.byKey {
		if mem.WorkspaceID() == workspaceID {
			out = append(out, mem)
		}
	}
	return out, nil
}

func (m *memMemberships) ListMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.WorkspaceMember, error) {
	_ = ctx
	out := make([]*domain.WorkspaceMember, 0)
	for _, mem := range m.members {
		if mem.WorkspaceID() == workspaceID {
			out = append(out, mem)
		}
	}
	return out, nil
}

func (m *memMemberships) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Membership, error) {
	_ = ctx
	out := make([]*domain.Membership, 0)
	for _, mem := range m.byKey {
		if mem.UserID() == userID {
			out = append(out, mem)
		}
	}
	return out, nil
}

func TestListWorkspaceMembers(t *testing.T) {
	t.Parallel()

	wsID := uuid.Must(uuid.NewV7())
	actorID := uuid.Must(uuid.NewV7())
	otherWS := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()

	memberA := domain.ReconstituteWorkspaceMember(
		wsID, uuid.Must(uuid.NewV7()), "a@example.com", "Alice", domain.WorkspaceRoleOwner, now,
	)
	memberB := domain.ReconstituteWorkspaceMember(
		wsID, uuid.Must(uuid.NewV7()), "b@example.com", "Bob", domain.WorkspaceRoleMember, now.Add(time.Second),
	)
	memberOther := domain.ReconstituteWorkspaceMember(
		otherWS, uuid.Must(uuid.NewV7()), "c@example.com", "Carol", domain.WorkspaceRoleMember, now,
	)

	memberships := newMemMemberships()
	memberships.members = []*domain.WorkspaceMember{memberA, memberB, memberOther}
	_ = memberships.Save(context.Background(), domain.ReconstituteMembership(wsID, actorID, domain.WorkspaceRoleAdmin, now))

	svc := &application.Services{Memberships: memberships}

	t.Run("returns workspace members for actor", func(t *testing.T) {
		t.Parallel()
		got, err := svc.ListWorkspaceMembers(context.Background(), wsID, actorID)
		if err != nil {
			t.Fatalf("ListWorkspaceMembers: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len=%d want 2", len(got))
		}
	})

	t.Run("forbidden when actor not a member", func(t *testing.T) {
		t.Parallel()
		_, err := svc.ListWorkspaceMembers(context.Background(), wsID, uuid.Must(uuid.NewV7()))
		if err == nil {
			t.Fatal("expected error")
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != apperr.ErrCodeNotFound {
			t.Fatalf("got %#v want not_found", err)
		}
	})

	t.Run("empty workspace", func(t *testing.T) {
		t.Parallel()
		emptyWS := uuid.Must(uuid.NewV7())
		emptyActor := uuid.Must(uuid.NewV7())
		mems := newMemMemberships()
		_ = mems.Save(context.Background(), domain.ReconstituteMembership(emptyWS, emptyActor, domain.WorkspaceRoleOwner, now))
		emptySvc := &application.Services{Memberships: mems}
		got, err := emptySvc.ListWorkspaceMembers(context.Background(), emptyWS, emptyActor)
		if err != nil {
			t.Fatalf("ListWorkspaceMembers: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("len=%d want 0", len(got))
		}
	})
}
