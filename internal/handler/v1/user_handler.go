package v1handler

import (
	"net/http"
	v1dto "shopping-cart/internal/dto/v1"
	v1service "shopping-cart/internal/service/v1"
	"shopping-cart/internal/utils"
	"shopping-cart/internal/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	service v1service.UserService
}

func NewUserHandler(service v1service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var params v1dto.GetUserParams

	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	users, total, err := h.service.GetAllUsers(c, params.Search, params.OrderBy, params.Sort, params.Page, params.Limit, false)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	dtos := v1dto.ToUserDTOs(users)

	paginationResponse := utils.NewPaginationResponse(dtos, params.Page, params.Limit, int32(total))

	utils.ResponseSuccess(c, http.StatusOK, "Users fetched successfully", paginationResponse)
}

func (h *UserHandler) GetSoftDeletedUsers(c *gin.Context) {
	var params v1dto.GetUserParams

	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	users, total, err := h.service.GetAllUsers(c, params.Search, params.OrderBy, params.Sort, params.Page, params.Limit, true)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	dtos := v1dto.ToUserDTOs(users)

	paginationResponse := utils.NewPaginationResponse(dtos, params.Page, params.Limit, int32(total))

	utils.ResponseSuccess(c, http.StatusOK, "Users fetched successfully", paginationResponse)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input v1dto.CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	params := input.ToUserModel() // Convert input to sqlc model

	// Create user
	user, err := h.service.CreateUser(c, params)

	// Handle error
	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	// Response success
	utils.ResponseSuccess(c, http.StatusCreated, "User created successfully", v1dto.ToUserDTO(user))
}

func (h *UserHandler) GetUserByUuid(c *gin.Context) {
	var param v1dto.GetUserByUuidParam

	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	userUuid, err := uuid.Parse(param.Uuid)

	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid user UUID", utils.ErrCodeBadRequest))
		return
	}

	user, err := h.service.GetUserByUuid(c, userUuid)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "User fetched successfully", v1dto.ToUserDTO(user))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var param v1dto.GetUserByUuidParam

	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	var input v1dto.UpdateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	userUuid, err := uuid.Parse(param.Uuid)

	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid user UUID", utils.ErrCodeBadRequest))
		return
	}

	userModel := input.ToUserModel(userUuid)

	user, err := h.service.UpdateUser(c, userModel)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "User updated successfully", v1dto.ToUserDTO(user))
}

func (h *UserHandler) SoftDeleteUser(c *gin.Context) {
	var param v1dto.GetUserByUuidParam

	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	userUuid, err := uuid.Parse(param.Uuid)

	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid user UUID", utils.ErrCodeBadRequest))
		return
	}

	user, err := h.service.SoftDeleteUser(c, userUuid)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "User deleted successfully", v1dto.ToUserDTO(user))
}

func (h *UserHandler) RestoreUser(c *gin.Context) {
	var param v1dto.GetUserByUuidParam

	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	userUuid, err := uuid.Parse(param.Uuid)

	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid user UUID", utils.ErrCodeBadRequest))
		return
	}

	user, err := h.service.RestoreUser(c, userUuid)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "User restored successfully", v1dto.ToUserDTO(user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	var param v1dto.GetUserByUuidParam

	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	userUuid, err := uuid.Parse(param.Uuid)

	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid user UUID", utils.ErrCodeBadRequest))
		return
	}

	err = h.service.DeleteUser(c, userUuid)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseStatusCode(c, http.StatusNoContent)
}
