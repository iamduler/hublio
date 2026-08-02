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

type memOrgs struct {
	byID map[uuid.UUID]*domain.Organization
}

func newMemOrgs() *memOrgs {
	return &memOrgs{byID: map[uuid.UUID]*domain.Organization{}}
}

func (m *memOrgs) Save(ctx context.Context, org *domain.Organization) error {
	_ = ctx
	m.byID[org.ID()] = org
	return nil
}

func (m *memOrgs) Update(ctx context.Context, org *domain.Organization) error {
	return m.Save(ctx, org)
}

func (m *memOrgs) FindByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	_ = ctx
	org, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return org, nil
}

func (m *memOrgs) FindByName(ctx context.Context, name string) (*domain.Organization, error) {
	_ = ctx
	for _, org := range m.byID {
		if org.Name() == name {
			return org, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *memOrgs) List(ctx context.Context) ([]*domain.Organization, error) {
	_ = ctx
	out := make([]*domain.Organization, 0, len(m.byID))
	for _, org := range m.byID {
		if org.DeletedAt() != nil {
			continue
		}
		out = append(out, org)
	}
	return out, nil
}

var _ domain.OrganizationRepository = (*memOrgs)(nil)

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	t.Run("returns user and organization", func(t *testing.T) {
		t.Parallel()
		orgID := uuid.Must(uuid.NewV7())
		userID := uuid.Must(uuid.NewV7())
		org := domain.ReconstituteOrganization(orgID, "Acme", domain.OrganizationStatusActive, now, now, nil)
		user := domain.ReconstituteUser(
			userID, orgID, "me@example.com", "Me User", "hash",
			true, false, domain.UserStatusActive,
			now, now, nil, nil, nil, nil,
		)
		users := newMemUsers()
		orgs := newMemOrgs()
		_ = users.Save(context.Background(), user)
		_ = orgs.Save(context.Background(), org)
		svc := &application.Services{Users: users, Orgs: orgs}

		gotUser, gotOrg, err := svc.GetCurrentUser(context.Background(), userID)
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}
		if gotUser.ID() != userID || gotUser.Email() != "me@example.com" {
			t.Fatalf("user: %+v", gotUser)
		}
		if gotOrg.ID() != orgID || gotOrg.Name() != "Acme" {
			t.Fatalf("org: %+v", gotOrg)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		t.Parallel()
		svc := &application.Services{Users: newMemUsers(), Orgs: newMemOrgs()}
		_, _, err := svc.GetCurrentUser(context.Background(), uuid.Must(uuid.NewV7()))
		if err == nil {
			t.Fatal("expected error")
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != apperr.ErrCodeUnauthorized {
			t.Fatalf("got %#v", err)
		}
	})

	t.Run("suspended user", func(t *testing.T) {
		t.Parallel()
		orgID := uuid.Must(uuid.NewV7())
		userID := uuid.Must(uuid.NewV7())
		org := domain.ReconstituteOrganization(orgID, "Acme", domain.OrganizationStatusActive, now, now, nil)
		user := domain.ReconstituteUser(
			userID, orgID, "suspended@example.com", "Suspended", "hash",
			false, false, domain.UserStatusSuspended,
			now, now, nil, nil, nil, nil,
		)
		users := newMemUsers()
		orgs := newMemOrgs()
		_ = users.Save(context.Background(), user)
		_ = orgs.Save(context.Background(), org)
		svc := &application.Services{Users: users, Orgs: orgs}

		_, _, err := svc.GetCurrentUser(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != apperr.ErrCodeUnauthorized {
			t.Fatalf("got %#v", err)
		}
	})
}
