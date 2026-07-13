package v1dto

import (
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/utils"

	"github.com/google/uuid"
)

type UserDTO struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       *int   `json:"age"`
	Status    string `json:"status"`
	Level     string `json:"level"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

type CreateUserInput struct {
	Name     string `json:"name" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Age      int32  `json:"age" binding:"omitempty,min_int=1"`
	Password string `json:"password" binding:"required,min=8,max=50,password_advanced"`
	Status   int32  `json:"status" binding:"required,oneof=1 2 3"`
	Level    int32  `json:"level" binding:"required,oneof=1 2 3"`
}

type UpdateUserInput struct {
	Name     *string `json:"name" binding:"omitempty,min=3,max=100"`
	Age      *int32  `json:"age" binding:"omitempty,min_int=18"`
	Password *string `json:"password" binding:"omitempty,min=8,max=50,password_advanced"`
	Status   *int32  `json:"status" binding:"omitempty,oneof=1 2 3"`
	Level    *int32  `json:"level" binding:"omitempty,oneof=1 2 3"`
}

type GetUserByUuidParam struct {
	Uuid string `uri:"uuid" binding:"required,uuid"`
}

type GetUserParams struct {
	Search  string `form:"search" binding:"omitempty,min=3,max=50,search"`
	Page    int32  `form:"page" binding:"omitempty,min_int=1"`
	Limit   int32  `form:"limit" binding:"omitempty,min_int=1,max_int=100"`
	OrderBy string `form:"order_by" binding:"omitempty,oneof=id created_at"`
	Sort    string `form:"sort" binding:"omitempty,oneof=asc desc"`
}

func (input *CreateUserInput) ToUserModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		Email:    input.Email,
		Password: input.Password,
		FullName: input.Name,
		Age:      utils.ConvertToInt32Pointer(input.Age),
		Status:   input.Status,
		Level:    input.Level,
	}
}

func (input *UpdateUserInput) ToUserModel(uuid uuid.UUID) sqlc.UpdateUserParams {
	return sqlc.UpdateUserParams{
		Password: input.Password,
		FullName: input.Name,
		Age:      input.Age,
		Status:   input.Status,
		Level:    input.Level,
		Uuid:     uuid,
	}
}

func ToUserDTO(user sqlc.User) *UserDTO {
	dto := &UserDTO{
		UUID:      user.Uuid.String(),
		Name:      user.FullName,
		Email:     user.Email,
		Status:    mapStatusText(int(user.Status)),
		Level:     mapLevelText(int(user.Level)),
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if user.Age != nil {
		age := int(*user.Age)
		dto.Age = &age
	}

	if user.DeletedAt.Valid {
		dto.DeletedAt = user.DeletedAt.Time.Format("2006-01-02 15:04:05")
	} else {
		dto.DeletedAt = ""
	}

	return dto
}

func ToUserDTOs(users []sqlc.User) []UserDTO {
	dtos := make([]UserDTO, 0, len(users))

	for _, user := range users {
		dtos = append(dtos, *ToUserDTO(user))
	}

	return dtos
}

func mapStatusText(status int) string {
	switch status {
	case 1:
		return "active"
	case 2:
		return "inactive"
	case 3:
		return "banned"
	default:
		return "unknown"
	}
}

func mapLevelText(level int) string {
	switch level {
	case 1:
		return "admin"
	case 2:
		return "moderator"
	case 3:
		return "member"
	default:
		return "unknown"
	}
}
