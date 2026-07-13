package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shopping-cart/internal/db"
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/utils"

	"github.com/google/uuid"
)

type SqlUserRepository struct {
	db sqlc.Querier
}

func NewSqlUserRepository(db sqlc.Querier) UserRepository {
	return &SqlUserRepository{
		db: db,
	}
}

func (r *SqlUserRepository) GetAll(ctx context.Context, search, orderBy, sort string, offset, limit int32) ([]sqlc.User, error) {
	var users []sqlc.User
	var err error

	switch {
	case orderBy == "id" && sort == "asc":
		users, err = r.db.ListUsersWUserIdAsc(ctx, sqlc.ListUsersWUserIdAscParams{
			Offset: offset,
			Limit:  limit,
			Search: &search,
		})
	case orderBy == "id" && sort == "desc":
		users, err = r.db.ListUsersWUserIdDesc(ctx, sqlc.ListUsersWUserIdDescParams{
			Offset: offset,
			Limit:  limit,
			Search: &search,
		})
	case orderBy == "created_at" && sort == "asc":
		users, err = r.db.ListUsersWCreatedAtAsc(ctx, sqlc.ListUsersWCreatedAtAscParams{
			Offset: offset,
			Limit:  limit,
			Search: &search,
		})
	case orderBy == "created_at" && sort == "desc":
		users, err = r.db.ListUsersWCreatedAtDesc(ctx, sqlc.ListUsersWCreatedAtDescParams{
			Offset: offset,
			Limit:  limit,
			Search: &search,
		})

		if err != nil {
			return []sqlc.User{}, err
		}
	}

	return users, nil
}

func (r *SqlUserRepository) GetAllV2(ctx context.Context, search, orderBy, sort string, offset, limit int32, isDeleted bool) ([]sqlc.User, error) {
	query := `SELECT * FROM users
		WHERE (
			$1::TEXT IS NULL
			OR $1::TEXT = ''
			OR email ILIKE '%' || $1 || '%'
			OR full_name ILIKE '%' || $1 || '%'
		)`

	if isDeleted {
		query += " AND deleted_at IS NOT NULL"
	} else {
		query += " AND deleted_at IS NULL"
	}

	order := "ASC"

	if sort == "desc" {
		order = "DESC"
	}

	switch orderBy {
	case "id", "created_at":
		query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
	default:
		query += " ORDER BY created_at DESC"
	}

	query += " LIMIT $2 OFFSET $3"

	rows, err := db.DBPool.Query(ctx, query, search, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items := []sqlc.User{}

	for rows.Next() {
		var i sqlc.User
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.Email,
			&i.Password,
			&i.FullName,
			&i.Age,
			&i.Status,
			&i.Level,
			&i.DeletedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SqlUserRepository) Count(ctx context.Context, search string, isDeleted bool) (int64, error) {
	count, err := r.db.CountUsers(ctx, sqlc.CountUsersParams{
		Deleted: &isDeleted,
		Search:  &search,
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SqlUserRepository) Create(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	user, err := r.db.CreateUser(ctx, params)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) GetByUuid(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := r.db.GetUser(ctx, uuid)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) Update(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error) {
	user, err := r.db.UpdateUser(ctx, params)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) SoftDelete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := r.db.SoftDeleteUser(ctx, uuid)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) Restore(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := r.db.RestoreUser(ctx, uuid)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) Delete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := r.db.TrashUser(ctx, uuid)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)

	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (r *SqlUserRepository) UpdatePassword(ctx context.Context, params sqlc.UpdatePasswordParams) (sqlc.User, error) {
	user, err := r.db.UpdatePassword(ctx, params)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("User not found", utils.ErrCodeNotFound)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to update password", utils.ErrCodeInternal)
	}

	return user, nil
}
