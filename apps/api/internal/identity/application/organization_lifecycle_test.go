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

func TestListOrganizations(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	t.Run("platform admin lists all", func(t *testing.T) {
		t.Parallel()
		adminOrgID := uuid.Must(uuid.NewV7())
		otherOrgID := uuid.Must(uuid.NewV7())
		adminID := uuid.Must(uuid.NewV7())

		adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
		otherOrg := domain.ReconstituteOrganization(otherOrgID, "Acme", domain.OrganizationStatusActive, now, now, nil)
		admin := domain.ReconstituteUser(
			adminID, adminOrgID, "admin@example.com", "Admin", "hash",
			true, true, domain.UserStatusActive,
			now, now, nil, nil, nil, nil,
		)

		users := newMemUsers()
		orgs := newMemOrgs()
		_ = users.Save(context.Background(), admin)
		_ = orgs.Save(context.Background(), adminOrg)
		_ = orgs.Save(context.Background(), otherOrg)
		svc := &application.Services{Users: users, Orgs: orgs}

		list, err := svc.ListOrganizations(context.Background(), adminID)
		if err != nil {
			t.Fatalf("ListOrganizations: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("want 2 orgs, got %d", len(list))
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		orgID := uuid.Must(uuid.NewV7())
		userID := uuid.Must(uuid.NewV7())
		org := domain.ReconstituteOrganization(orgID, "Acme", domain.OrganizationStatusActive, now, now, nil)
		user := domain.ReconstituteUser(
			userID, orgID, "user@example.com", "User", "hash",
			true, false, domain.UserStatusActive,
			now, now, nil, nil, nil, nil,
		)
		users := newMemUsers()
		orgs := newMemOrgs()
		_ = users.Save(context.Background(), user)
		_ = orgs.Save(context.Background(), org)
		svc := &application.Services{Users: users, Orgs: orgs}

		_, err := svc.ListOrganizations(context.Background(), userID)
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != apperr.ErrCodeForbidden {
			t.Fatalf("got %#v", err)
		}
	})
}

func TestSuspendActivateOrganizationAsPlatformAdmin(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	adminOrgID := uuid.Must(uuid.NewV7())
	targetOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())

	adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
	targetOrg := domain.ReconstituteOrganization(targetOrgID, "Tenant", domain.OrganizationStatusActive, now, now, nil)
	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)

	users := newMemUsers()
	orgs := newMemOrgs()
	_ = users.Save(context.Background(), admin)
	_ = orgs.Save(context.Background(), adminOrg)
	_ = orgs.Save(context.Background(), targetOrg)
	svc := &application.Services{Users: users, Orgs: orgs}

	suspended, err := svc.SuspendOrganization(context.Background(), targetOrgID, adminID)
	if err != nil {
		t.Fatalf("SuspendOrganization: %v", err)
	}
	if suspended.Status() != domain.OrganizationStatusSuspended {
		t.Fatalf("status = %s", suspended.Status())
	}

	activated, err := svc.ActivateOrganization(context.Background(), targetOrgID, adminID)
	if err != nil {
		t.Fatalf("ActivateOrganization: %v", err)
	}
	if activated.Status() != domain.OrganizationStatusActive {
		t.Fatalf("status = %s", activated.Status())
	}
}

func TestSuspendOrganizationSameOrgMember(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	orgID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	otherOrgID := uuid.Must(uuid.NewV7())

	org := domain.ReconstituteOrganization(orgID, "Acme", domain.OrganizationStatusActive, now, now, nil)
	other := domain.ReconstituteOrganization(otherOrgID, "Other", domain.OrganizationStatusActive, now, now, nil)
	user := domain.ReconstituteUser(
		userID, orgID, "owner@example.com", "Owner", "hash",
		true, false, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)

	users := newMemUsers()
	orgs := newMemOrgs()
	_ = users.Save(context.Background(), user)
	_ = orgs.Save(context.Background(), org)
	_ = orgs.Save(context.Background(), other)
	svc := &application.Services{Users: users, Orgs: orgs}

	got, err := svc.SuspendOrganization(context.Background(), orgID, userID)
	if err != nil {
		t.Fatalf("same-org suspend: %v", err)
	}
	if got.Status() != domain.OrganizationStatusSuspended {
		t.Fatalf("status = %s", got.Status())
	}

	_, err = svc.SuspendOrganization(context.Background(), otherOrgID, userID)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != apperr.ErrCodeForbidden {
		t.Fatalf("cross-org got %#v", err)
	}
}
