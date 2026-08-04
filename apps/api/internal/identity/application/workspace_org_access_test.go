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

type memWorkspaces struct {
	byID map[uuid.UUID]*domain.Workspace
}

func newMemWorkspaces() *memWorkspaces {
	return &memWorkspaces{byID: map[uuid.UUID]*domain.Workspace{}}
}

func (m *memWorkspaces) Save(ctx context.Context, ws *domain.Workspace) error {
	_ = ctx
	m.byID[ws.ID()] = ws
	return nil
}

func (m *memWorkspaces) Update(ctx context.Context, ws *domain.Workspace) error {
	return m.Save(ctx, ws)
}

func (m *memWorkspaces) FindByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	_ = ctx
	ws, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return ws, nil
}

func (m *memWorkspaces) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.Workspace, error) {
	_ = ctx
	out := make([]*domain.Workspace, 0)
	for _, ws := range m.byID {
		if ws.OrganizationID() == organizationID && ws.DeletedAt() == nil {
			out = append(out, ws)
		}
	}
	return out, nil
}

var _ domain.WorkspaceRepository = (*memWorkspaces)(nil)

func TestListWorkspacesPlatformAdminCrossTenant(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	adminOrgID := uuid.Must(uuid.NewV7())
	tenantOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())
	wsID := uuid.Must(uuid.NewV7())

	adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
	tenantOrg := domain.ReconstituteOrganization(tenantOrgID, "Tenant", domain.OrganizationStatusActive, now, now, nil)
	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	ws := domain.ReconstituteWorkspace(wsID, tenantOrgID, "prod", "production", domain.WorkspaceStatusActive, now, now, nil)

	users := newMemUsers()
	orgs := newMemOrgs()
	workspaces := newMemWorkspaces()
	_ = users.Save(context.Background(), admin)
	_ = orgs.Save(context.Background(), adminOrg)
	_ = orgs.Save(context.Background(), tenantOrg)
	_ = workspaces.Save(context.Background(), ws)
	svc := &application.Services{Users: users, Orgs: orgs, Workspaces: workspaces}

	list, err := svc.ListWorkspaces(context.Background(), tenantOrgID, adminID)
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 workspace, got %d", len(list))
	}
}

func TestListWorkspacesNonAdminCrossTenantForbidden(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	orgA := uuid.Must(uuid.NewV7())
	orgB := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := domain.ReconstituteUser(
		userID, orgA, "user@example.com", "User", "hash",
		true, false, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	users := newMemUsers()
	_ = users.Save(context.Background(), user)
	svc := &application.Services{Users: users, Workspaces: newMemWorkspaces()}

	_, err := svc.ListWorkspaces(context.Background(), orgB, userID)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != apperr.ErrCodeForbidden {
		t.Fatalf("got %#v", err)
	}
}

func TestCreateWorkspacePlatformAdminNoMembership(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	adminOrgID := uuid.Must(uuid.NewV7())
	tenantOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())

	adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
	tenantOrg := domain.ReconstituteOrganization(tenantOrgID, "Tenant", domain.OrganizationStatusActive, now, now, nil)
	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)

	users := newMemUsers()
	orgs := newMemOrgs()
	workspaces := newMemWorkspaces()
	memberships := newMemMemberships()
	_ = users.Save(context.Background(), admin)
	_ = orgs.Save(context.Background(), adminOrg)
	_ = orgs.Save(context.Background(), tenantOrg)
	svc := &application.Services{
		Users: users, Orgs: orgs, Workspaces: workspaces, Memberships: memberships,
	}

	ws, err := svc.CreateWorkspace(context.Background(), application.CreateWorkspaceInput{
		OrganizationID: tenantOrgID,
		ActorUserID:    adminID,
		Name:           "ops",
		Environment:    "staging",
	})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if ws.OrganizationID() != tenantOrgID {
		t.Fatalf("org mismatch")
	}
	if len(memberships.byKey) != 0 {
		t.Fatalf("platform admin must not auto-join membership, got %d", len(memberships.byKey))
	}
}

func TestSetWorkspaceStatusPlatformAdmin(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	adminOrgID := uuid.Must(uuid.NewV7())
	tenantOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())
	wsID := uuid.Must(uuid.NewV7())

	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	ws := domain.ReconstituteWorkspace(wsID, tenantOrgID, "prod", "production", domain.WorkspaceStatusActive, now, now, nil)

	users := newMemUsers()
	workspaces := newMemWorkspaces()
	_ = users.Save(context.Background(), admin)
	_ = workspaces.Save(context.Background(), ws)
	svc := &application.Services{Users: users, Workspaces: workspaces, Memberships: newMemMemberships()}

	disabled, err := svc.SetWorkspaceStatus(context.Background(), wsID, adminID, false)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if disabled.Status() != domain.WorkspaceStatusDisabled {
		t.Fatalf("status = %s", disabled.Status())
	}
}

func TestCreateOrganizationPlatformAdmin(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	adminOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())

	adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)

	users := newMemUsers()
	orgs := newMemOrgs()
	workspaces := newMemWorkspaces()
	_ = users.Save(context.Background(), admin)
	_ = orgs.Save(context.Background(), adminOrg)
	svc := &application.Services{Users: users, Orgs: orgs, Workspaces: workspaces}

	result, err := svc.CreateOrganization(context.Background(), application.CreateOrganizationInput{
		ActorUserID: adminID,
		Name:        "New Tenant",
	})
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if result.Organization.Name() != "New Tenant" {
		t.Fatalf("name = %s", result.Organization.Name())
	}
	if result.Workspace.Name() != "default" {
		t.Fatalf("workspace = %s", result.Workspace.Name())
	}
}

func TestCreateOrganizationNonAdminForbidden(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	orgID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	user := domain.ReconstituteUser(
		userID, orgID, "user@example.com", "User", "hash",
		true, false, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	users := newMemUsers()
	_ = users.Save(context.Background(), user)
	svc := &application.Services{Users: users, Orgs: newMemOrgs(), Workspaces: newMemWorkspaces()}

	_, err := svc.CreateOrganization(context.Background(), application.CreateOrganizationInput{
		ActorUserID: userID,
		Name:        "Nope",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != apperr.ErrCodeForbidden {
		t.Fatalf("got %#v", err)
	}
}

func TestUpdateAndArchiveOrganization(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	adminOrgID := uuid.Must(uuid.NewV7())
	targetOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())

	adminOrg := domain.ReconstituteOrganization(adminOrgID, "Platform", domain.OrganizationStatusActive, now, now, nil)
	targetOrg := domain.ReconstituteOrganization(targetOrgID, "Old Name", domain.OrganizationStatusActive, now, now, nil)
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

	updated, err := svc.UpdateOrganization(context.Background(), targetOrgID, adminID, "New Name")
	if err != nil {
		t.Fatalf("UpdateOrganization: %v", err)
	}
	if updated.Name() != "New Name" {
		t.Fatalf("name = %s", updated.Name())
	}

	archived, err := svc.ArchiveOrganization(context.Background(), targetOrgID, adminID)
	if err != nil {
		t.Fatalf("ArchiveOrganization: %v", err)
	}
	if archived.Status() != domain.OrganizationStatusArchived || archived.DeletedAt() == nil {
		t.Fatalf("archive status = %s deleted=%v", archived.Status(), archived.DeletedAt())
	}
}

func TestListOrganizationUsersPlatformAdmin(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	adminOrgID := uuid.Must(uuid.NewV7())
	tenantOrgID := uuid.Must(uuid.NewV7())
	adminID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	admin := domain.ReconstituteUser(
		adminID, adminOrgID, "admin@example.com", "Admin", "hash",
		true, true, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	tenantUser := domain.ReconstituteUser(
		userID, tenantOrgID, "alice@tenant.com", "Alice", "hash",
		true, false, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)

	users := newMemUsers()
	_ = users.Save(context.Background(), admin)
	_ = users.Save(context.Background(), tenantUser)
	svc := &application.Services{Users: users}

	list, err := svc.ListOrganizationUsers(context.Background(), tenantOrgID, adminID)
	if err != nil {
		t.Fatalf("ListOrganizationUsers: %v", err)
	}
	if len(list) != 1 || list[0].Email() != "alice@tenant.com" {
		t.Fatalf("got %#v", list)
	}
}

func TestListOrganizationUsersNonAdminCrossTenantForbidden(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	orgA := uuid.Must(uuid.NewV7())
	orgB := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := domain.ReconstituteUser(
		userID, orgA, "user@example.com", "User", "hash",
		true, false, domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	users := newMemUsers()
	_ = users.Save(context.Background(), user)
	svc := &application.Services{Users: users}

	_, err := svc.ListOrganizationUsers(context.Background(), orgB, userID)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != apperr.ErrCodeForbidden {
		t.Fatalf("got %#v", err)
	}
}
