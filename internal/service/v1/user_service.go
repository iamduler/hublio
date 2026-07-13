package v1service

import (
	"database/sql"
	"errors"
	"fmt"
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/repository"
	"shopping-cart/internal/utils"
	"strings"
	"time"

	"shopping-cart/pkg/cache"
	"shopping-cart/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo  repository.UserRepository
	cache cache.RedisCacheService
}

func NewUserService(repo repository.UserRepository, redis *redis.Client) UserService {
	return &userService{
		repo:  repo,
		cache: cache.NewRedisCacheService(redis),
	}
}

func (s *userService) GetAllUsers(ctx *gin.Context, search, orderBy, sort string, page, limit int32, isDeleted bool) ([]sqlc.User, int64, error) {
	context := ctx.Request.Context() // Get the context from the request

	if sort == "" {
		sort = "desc"
	}

	if orderBy == "" {
		orderBy = "created_at"
	}

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		envLimit := utils.GetIntEnv("LIMIT_ITEMS_PER_PAGE", 10)
		limit = int32(envLimit)
	}

	offset := (page - 1) * limit

	// Check cache
	cacheKey := s.generateCacheKey(search, orderBy, sort, page, limit, isDeleted)
	var cacheData struct {
		Users []sqlc.User `json:"users"`
		Total int64       `json:"total"`
	}

	err := s.cache.Get(cacheKey, &cacheData)

	if err == nil && cacheData.Users != nil {
		return cacheData.Users, cacheData.Total, nil
	}

	users, err := s.repo.GetAllV2(context, search, orderBy, sort, offset, limit, isDeleted)

	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(err, "Failed to get users", utils.ErrCodeInternal)
	}

	total, err := s.repo.Count(context, search, isDeleted)

	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(err, "Failed to get total users", utils.ErrCodeInternal)
	}

	// Set cache
	cacheData = struct {
		Users []sqlc.User `json:"users"`
		Total int64       `json:"total"`
	}{
		Users: users,
		Total: total,
	}

	s.cache.Set(cacheKey, cacheData, 5*time.Minute)

	return users, total, nil
}

func (s *userService) CreateUser(ctx *gin.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context() // Get the context from the request

	// Validate params
	params.Email = utils.NormalizeString(params.Email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)

	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "Failed to hash password", utils.ErrCodeInternal)
	}

	// Hash password
	params.Password = string(hashedPassword)

	// Create user
	user, err := s.repo.Create(context, params)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, utils.NewError("Email already exists", utils.ErrCodeConflict)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to create user", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cache.Delete("users:*"); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return user, nil
}

func (s *userService) GetUserByUuid(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context() // Get the context from the request

	user, err := s.repo.GetByUuid(context, uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("User not found", utils.ErrCodeNotFound)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to get user", utils.ErrCodeInternal)
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx *gin.Context, params sqlc.UpdateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context() // Get the context from the request

	if params.Password != nil && *params.Password != "" {
		// Validate params
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*params.Password), bcrypt.DefaultCost)

		if err != nil {
			return sqlc.User{}, utils.WrapError(err, "Failed to hash password", utils.ErrCodeInternal)
		}

		hashed := string(hashedPassword)
		params.Password = &hashed
	}

	// Update user
	user, err := s.repo.Update(context, params)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("User not found", utils.ErrCodeNotFound)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to update user", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cache.Delete("users:*"); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return user, nil
}

func (s *userService) SoftDeleteUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context() // Get the context from the request

	user, err := s.repo.SoftDelete(context, uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("User not found or already deleted", utils.ErrCodeNotFound)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to delete user", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cache.Delete("users:*"); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return user, nil
}

func (s *userService) RestoreUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context() // Get the context from the request

	user, err := s.repo.Restore(context, uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("User not found or already restored", utils.ErrCodeNotFound)
		}

		return sqlc.User{}, utils.WrapError(err, "Failed to restore user", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cache.Delete("users:*"); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx *gin.Context, uuid uuid.UUID) error {
	context := ctx.Request.Context() // Get the context from the request

	_, err := s.repo.Delete(context, uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NewError("User not found or already deleted", utils.ErrCodeNotFound)
		}

		return utils.WrapError(err, "Failed to delete user", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cache.Delete("users:*"); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return nil
}

func (s *userService) generateCacheKey(search, orderBy, sort string, page, limit int32, isDeleted bool) string {
	search = strings.TrimSpace(search)

	// Search
	if search == "" {
		search = "none"
	}

	// Order by
	orderBy = strings.TrimSpace(orderBy)

	if orderBy == "" {
		orderBy = "created_at"
	}

	// Sort
	sort = strings.ToLower(strings.TrimSpace(sort))

	if sort == "" {
		sort = "desc"
	}

	// Page
	if page <= 0 {
		page = 1
	}

	// Limit
	if limit <= 0 {
		envLimit := utils.GetIntEnv("LIMIT_ITEMS_PER_PAGE", 10)
		limit = int32(envLimit)
	}

	// Cache key
	return fmt.Sprintf("users:%s:%s:%s:%d:%d:%t", search, orderBy, sort, page, limit, isDeleted)
}
