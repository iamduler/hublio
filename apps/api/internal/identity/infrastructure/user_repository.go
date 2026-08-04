package infrastructure

import (
	"context"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func passwordHashPtr(hash string) *string {
	if hash == "" {
		return nil
	}
	h := hash
	return &h
}

func passwordHashValue(hash *string) string {
	if hash == nil {
		return ""
	}
	return *hash
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	return mapUnique(r.q(ctx).InsertUser(ctx, sqlc.InsertUserParams{
		ID:                user.ID(),
		OrganizationID:    user.OrganizationID(),
		Email:             user.Email(),
		FullName:          user.FullName(),
		IsActive:          user.IsActive(),
		IsPlatformAdmin:   user.IsPlatformAdmin(),
		PasswordHash:      passwordHashPtr(user.PasswordHash()),
		EmailVerifiedAt:   timestamptzPtr(user.EmailVerifiedAt()),
		PasswordChangedAt: timestamptzPtr(user.PasswordChangedAt()),
		LastLoginAt:       timestamptzPtr(user.LastLoginAt()),
		Status:            string(user.Status()),
		CreatedAt:         timestamptz(user.CreatedAt()),
		UpdatedAt:         timestamptz(user.UpdatedAt()),
		DeletedAt:         timestamptzPtr(user.DeletedAt()),
	}))
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return mapUnique(r.q(ctx).UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:                user.ID(),
		FullName:          user.FullName(),
		IsActive:          user.IsActive(),
		PasswordHash:      passwordHashPtr(user.PasswordHash()),
		EmailVerifiedAt:   timestamptzPtr(user.EmailVerifiedAt()),
		PasswordChangedAt: timestamptzPtr(user.PasswordChangedAt()),
		LastLoginAt:       timestamptzPtr(user.LastLoginAt()),
		Status:            string(user.Status()),
		UpdatedAt:         timestamptz(user.UpdatedAt()),
		DeletedAt:         timestamptzPtr(user.DeletedAt()),
	}))
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q(ctx).GetUserByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return reconstituteUser(
		row.ID,
		row.OrganizationID,
		row.Email,
		row.FullName,
		row.PasswordHash,
		row.IsActive,
		row.IsPlatformAdmin,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.LastLoginAt,
		row.DeletedAt,
		row.EmailVerifiedAt,
		row.PasswordChangedAt,
	), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return reconstituteUser(
		row.ID,
		row.OrganizationID,
		row.Email,
		row.FullName,
		row.PasswordHash,
		row.IsActive,
		row.IsPlatformAdmin,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
		row.LastLoginAt,
		row.DeletedAt,
		row.EmailVerifiedAt,
		row.PasswordChangedAt,
	), nil
}

func (r *UserRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.User, error) {
	rows, err := r.q(ctx).ListUsersByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		out = append(out, reconstituteUser(
			row.ID,
			row.OrganizationID,
			row.Email,
			row.FullName,
			row.PasswordHash,
			row.IsActive,
			row.IsPlatformAdmin,
			row.Status,
			row.CreatedAt,
			row.UpdatedAt,
			row.LastLoginAt,
			row.DeletedAt,
			row.EmailVerifiedAt,
			row.PasswordChangedAt,
		))
	}
	return out, nil
}

func reconstituteUser(
	id, organizationID uuid.UUID,
	email, fullName string,
	passwordHash *string,
	isActive, isPlatformAdmin bool,
	status string,
	createdAt, updatedAt, lastLoginAt, deletedAt, emailVerifiedAt, passwordChangedAt pgtype.Timestamptz,
) *domain.User {
	return domain.ReconstituteUser(
		id,
		organizationID,
		email,
		fullName,
		passwordHashValue(passwordHash),
		isActive,
		isPlatformAdmin,
		domain.UserStatus(status),
		timeFrom(createdAt),
		timeFrom(updatedAt),
		timePtrFrom(lastLoginAt),
		timePtrFrom(deletedAt),
		timePtrFrom(emailVerifiedAt),
		timePtrFrom(passwordChangedAt),
	)
}
