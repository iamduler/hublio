package infrastructure

import (
	"context"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	return mapUnique(r.q(ctx).InsertUser(ctx, sqlc.InsertUserParams{
		ID:                user.ID(),
		OrganizationID:    user.OrganizationID(),
		Email:             user.Email(),
		FullName:          user.FullName(),
		IsActive:          user.IsActive(),
		PasswordHash:      user.PasswordHash(),
		EmailVerifiedAt:   timestamptzPtr(nil),
		PasswordChangedAt: timestamptzPtr(nil),
		LastLoginAt:       timestamptzPtr(user.LastLoginAt()),
		Status:            string(user.Status()),
		CreatedAt:         timestamptz(user.CreatedAt()),
		UpdatedAt:         timestamptz(user.UpdatedAt()),
		DeletedAt:         timestamptzPtr(user.DeletedAt()),
	}))
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return mapUnique(r.q(ctx).UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:           user.ID(),
		FullName:     user.FullName(),
		IsActive:     user.IsActive(),
		PasswordHash: user.PasswordHash(),
		LastLoginAt:  timestamptzPtr(user.LastLoginAt()),
		Status:       string(user.Status()),
		UpdatedAt:    timestamptz(user.UpdatedAt()),
		DeletedAt:    timestamptzPtr(user.DeletedAt()),
	}))
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q(ctx).GetUserByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapUser(row), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapUser(row), nil
}

func mapUser(row sqlc.User) *domain.User {
	return domain.ReconstituteUser(
		row.ID,
		row.OrganizationID,
		row.Email,
		row.FullName,
		row.PasswordHash,
		row.IsActive,
		domain.UserStatus(row.Status),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.LastLoginAt),
		timePtrFrom(row.DeletedAt),
	)
}
