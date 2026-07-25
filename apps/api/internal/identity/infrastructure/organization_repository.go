package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

func (r *OrganizationRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *OrganizationRepository) Save(ctx context.Context, org *domain.Organization) error {
	err := r.q(ctx).InsertOrganization(ctx, sqlc.InsertOrganizationParams{
		ID:        org.ID(),
		Name:      org.Name(),
		Status:    string(org.Status()),
		CreatedAt: timestamptz(org.CreatedAt()),
		UpdatedAt: timestamptz(org.UpdatedAt()),
		DeletedAt: timestamptzPtr(org.DeletedAt()),
	})
	return mapUnique(err)
}

func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	return mapUnique(r.q(ctx).UpdateOrganization(ctx, sqlc.UpdateOrganizationParams{
		ID:        org.ID(),
		Name:      org.Name(),
		Status:    string(org.Status()),
		UpdatedAt: timestamptz(org.UpdatedAt()),
		DeletedAt: timestamptzPtr(org.DeletedAt()),
	}))
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	row, err := r.q(ctx).GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapOrganization(row), nil
}

func (r *OrganizationRepository) FindByName(ctx context.Context, name string) (*domain.Organization, error) {
	row, err := r.q(ctx).GetOrganizationByName(ctx, name)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapOrganization(row), nil
}

func mapOrganization(row sqlc.Organization) *domain.Organization {
	return domain.ReconstituteOrganization(
		row.ID,
		row.Name,
		domain.OrganizationStatus(row.Status),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.DeletedAt),
	)
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("identity repo: %w", err)
}

func mapUnique(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return fmt.Errorf("identity repo: %w", err)
}
