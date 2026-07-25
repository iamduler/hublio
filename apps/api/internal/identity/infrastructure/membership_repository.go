package infrastructure

import (
	"context"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepository struct {
	pool *pgxpool.Pool
}

func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

func (r *MembershipRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *MembershipRepository) Save(ctx context.Context, membership *domain.Membership) error {
	return mapUnique(r.q(ctx).InsertWorkspaceUser(ctx, sqlc.InsertWorkspaceUserParams{
		WorkspaceID: membership.WorkspaceID(),
		UserID:      membership.UserID(),
		Role:        string(membership.Role()),
		CreatedAt:   timestamptz(membership.CreatedAt()),
	}))
}

func (r *MembershipRepository) Find(ctx context.Context, workspaceID, userID uuid.UUID) (*domain.Membership, error) {
	row, err := r.q(ctx).GetWorkspaceUser(ctx, sqlc.GetWorkspaceUserParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapMembership(row), nil
}

func (r *MembershipRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Membership, error) {
	rows, err := r.q(ctx).ListWorkspaceUsersByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapMembership(row))
	}
	return out, nil
}

func (r *MembershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Membership, error) {
	rows, err := r.q(ctx).ListWorkspaceUsersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Membership, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapMembership(row))
	}
	return out, nil
}

func mapMembership(row sqlc.WorkspaceUser) *domain.Membership {
	return domain.ReconstituteMembership(
		row.WorkspaceID,
		row.UserID,
		domain.WorkspaceRole(row.Role),
		timeFrom(row.CreatedAt),
	)
}
