package repository

import (
	"context"
	"shopping-cart/internal/db/sqlc"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetAll(ctx context.Context, search, orderBy, sort string, offset, limit int32) ([]sqlc.User, error)
	GetAllV2(ctx context.Context, search, orderBy, sort string, offset, limit int32, isDeleted bool) ([]sqlc.User, error)
	Count(ctx context.Context, search string, isDeleted bool) (int64, error)
	Create(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	GetByUuid(ctx context.Context, uuid uuid.UUID) (sqlc.User, error)
	Update(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error)
	SoftDelete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error)
	Restore(ctx context.Context, uuid uuid.UUID) (sqlc.User, error)
	Delete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	UpdatePassword(ctx context.Context, params sqlc.UpdatePasswordParams) (sqlc.User, error)
}
