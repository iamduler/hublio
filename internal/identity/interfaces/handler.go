package interfaces

import (
	"context"
	"net/http"
	"strings"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	svc    *application.Services
	pool   *pgxpool.Pool
	tokens auth.TokenService
}

func NewHandler(svc *application.Services, pool *pgxpool.Pool, tokens auth.TokenService) *Handler {
	return &Handler{svc: svc, pool: pool, tokens: tokens}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, jwtAuth gin.HandlerFunc) {
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", h.register)
		authGroup.POST("/login", h.login)
		authGroup.POST("/logout", h.logout)
	}

	identity := api.Group("/identity")
	identity.Use(jwtAuth)
	{
		identity.GET("/organizations/:organizationId", h.getOrganization)
		identity.POST("/organizations/:organizationId/suspend", h.suspendOrganization)
		identity.POST("/organizations/:organizationId/activate", h.activateOrganization)

		identity.GET("/organizations/:organizationId/workspaces", h.listWorkspaces)
		identity.POST("/organizations/:organizationId/workspaces", h.createWorkspace)

		identity.POST("/workspaces/:workspaceId/enable", h.enableWorkspace)
		identity.POST("/workspaces/:workspaceId/disable", h.disableWorkspace)
		identity.POST("/workspaces/:workspaceId/members", h.addMember)

		identity.GET("/workspaces/:workspaceId/api-keys", h.listAPIKeys)
		identity.POST("/workspaces/:workspaceId/api-keys", h.createAPIKey)
		identity.POST("/workspaces/:workspaceId/api-keys/:apiKeyId/disable", h.disableAPIKey)
		identity.POST("/workspaces/:workspaceId/api-keys/:apiKeyId/rotate", h.rotateAPIKey)
	}
}

type registerRequest struct {
	OrganizationName string `json:"organization_name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	FullName         string `json:"full_name" binding:"required"`
	WorkspaceName    string `json:"workspace_name"`
	Environment      string `json:"environment"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var result *application.RegisterResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.Register(ctx, application.RegisterInput{
			OrganizationName: req.OrganizationName,
			Email:            req.Email,
			Password:         req.Password,
			FullName:         req.FullName,
			WorkspaceName:    req.WorkspaceName,
			Environment:      req.Environment,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	h.svc.PublishAfterCommit(c.Request.Context(),
		appendMany(
			result.Organization.PullEvents(),
			result.User.PullEvents(),
			result.Workspace.PullEvents(),
			result.Membership.PullEvents(),
		)...,
	)

	httpx.ResponseSuccess(c, http.StatusCreated, "registered", gin.H{
		"organization": organizationDTO(result.Organization),
		"workspace":    workspaceDTO(result.Workspace),
		"user":         userDTO(result.User),
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var result *application.LoginResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.Login(ctx, h.tokens, application.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	httpx.ResponseSuccess(c, http.StatusOK, "logged in", gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          userDTO(result.User),
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	if err := h.svc.Logout(c.Request.Context(), h.tokens, req.RefreshToken); err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "logged out", nil)
}

func (h *Handler) getOrganization(c *gin.Context) {
	orgID, ok := parseUUIDParam(c, "organizationId")
	if !ok {
		return
	}
	if !actorBelongsToOrg(c, orgID) {
		return
	}
	org, err := h.svc.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "organization", organizationDTO(org))
}

func (h *Handler) suspendOrganization(c *gin.Context) {
	h.orgLifecycle(c, h.svc.SuspendOrganization)
}

func (h *Handler) activateOrganization(c *gin.Context) {
	h.orgLifecycle(c, h.svc.ActivateOrganization)
}

func (h *Handler) orgLifecycle(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) (*domain.Organization, error)) {
	orgID, ok := parseUUIDParam(c, "organizationId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var org *domain.Organization
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		org, innerErr = fn(ctx, orgID, actorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), org.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, "organization updated", organizationDTO(org))
}

type createWorkspaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}

func (h *Handler) createWorkspace(c *gin.Context) {
	orgID, ok := parseUUIDParam(c, "organizationId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var req createWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	var ws *domain.Workspace
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		ws, innerErr = h.svc.CreateWorkspace(ctx, application.CreateWorkspaceInput{
			OrganizationID: orgID,
			ActorUserID:    actorID,
			Name:           req.Name,
			Environment:    req.Environment,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), ws.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusCreated, "workspace created", workspaceDTO(ws))
}

func (h *Handler) listWorkspaces(c *gin.Context) {
	orgID, ok := parseUUIDParam(c, "organizationId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	list, err := h.svc.ListWorkspaces(c.Request.Context(), orgID, actorID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, ws := range list {
		out = append(out, workspaceDTO(ws))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "workspaces", out)
}

func (h *Handler) enableWorkspace(c *gin.Context) {
	h.workspaceStatus(c, true)
}

func (h *Handler) disableWorkspace(c *gin.Context) {
	h.workspaceStatus(c, false)
}

func (h *Handler) workspaceStatus(c *gin.Context, enable bool) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var ws *domain.Workspace
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		ws, innerErr = h.svc.SetWorkspaceStatus(ctx, wsID, actorID, enable)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), ws.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, "workspace updated", workspaceDTO(ws))
}

type addMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

func (h *Handler) addMember(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	role, err := domain.ParseWorkspaceRole(req.Role)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid role", apperr.ErrCodeBadRequest))
		return
	}
	var mem *domain.Membership
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		mem, innerErr = h.svc.AddUserToWorkspace(ctx, application.AddMemberInput{
			WorkspaceID: wsID,
			ActorUserID: actorID,
			Email:       req.Email,
			Role:        role,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), mem.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusCreated, "member added", gin.H{
		"workspace_id": mem.WorkspaceID().String(),
		"user_id":      mem.UserID().String(),
		"role":         string(mem.Role()),
	})
}

type createAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) createAPIKey(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	var result *application.CreateAPIKeyResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.CreateAPIKey(ctx, application.CreateAPIKeyInput{
			WorkspaceID: wsID,
			ActorUserID: actorID,
			Name:        req.Name,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), result.APIKey.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusCreated, "api key created", gin.H{
		"api_key":   apiKeyDTO(result.APIKey),
		"plaintext": result.Plaintext,
		"warning":   "store plaintext now; it will not be shown again",
	})
}

func (h *Handler) listAPIKeys(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	keys, err := h.svc.ListAPIKeys(c.Request.Context(), wsID, actorID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(keys))
	for _, k := range keys {
		out = append(out, apiKeyDTO(k))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "api keys", out)
}

func (h *Handler) disableAPIKey(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	keyID, ok := parseUUIDParam(c, "apiKeyId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var key *domain.APIKey
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		key, innerErr = h.svc.DisableAPIKey(ctx, wsID, keyID, actorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), key.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, "api key disabled", apiKeyDTO(key))
}

func (h *Handler) rotateAPIKey(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	keyID, ok := parseUUIDParam(c, "apiKeyId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var result *application.CreateAPIKeyResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.RotateAPIKey(ctx, wsID, keyID, actorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), result.APIKey.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, "api key rotated", gin.H{
		"api_key":   apiKeyDTO(result.APIKey),
		"plaintext": result.Plaintext,
		"warning":   "store plaintext now; it will not be shown again",
	})
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid "+name, apperr.ErrCodeBadRequest))
		return uuid.Nil, false
	}
	return id, true
}

func actorUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, _ := c.Get("user_id")
	s, _ := raw.(string)
	id, err := uuid.Parse(s)
	if err != nil {
		httpx.ResponseError(c, apperr.New("unauthorized", apperr.ErrCodeUnauthorized))
		return uuid.Nil, false
	}
	return id, true
}

func actorBelongsToOrg(c *gin.Context, orgID uuid.UUID) bool {
	raw, _ := c.Get("organization_id")
	s, _ := raw.(string)
	if strings.TrimSpace(s) == "" || s != orgID.String() {
		httpx.ResponseError(c, apperr.New("forbidden", apperr.ErrCodeForbidden))
		return false
	}
	return true
}

func organizationDTO(o *domain.Organization) gin.H {
	return gin.H{
		"id":         o.ID().String(),
		"name":       o.Name(),
		"status":     string(o.Status()),
		"created_at": o.CreatedAt(),
		"updated_at": o.UpdatedAt(),
	}
}

func workspaceDTO(w *domain.Workspace) gin.H {
	return gin.H{
		"id":              w.ID().String(),
		"organization_id": w.OrganizationID().String(),
		"name":            w.Name(),
		"environment":     w.Environment(),
		"status":          string(w.Status()),
		"created_at":      w.CreatedAt(),
		"updated_at":      w.UpdatedAt(),
	}
}

func userDTO(u *domain.User) gin.H {
	return gin.H{
		"id":              u.ID().String(),
		"organization_id": u.OrganizationID().String(),
		"email":           u.Email(),
		"full_name":       u.FullName(),
		"status":          string(u.Status()),
	}
}

func apiKeyDTO(k *domain.APIKey) gin.H {
	return gin.H{
		"id":           k.ID().String(),
		"workspace_id": k.WorkspaceID().String(),
		"name":         k.Name(),
		"prefix":       k.Prefix(),
		"status":       string(k.Status()),
		"expires_at":   k.ExpiresAt(),
		"created_at":   k.CreatedAt(),
		"updated_at":   k.UpdatedAt(),
	}
}

func appendMany(batches ...[]domain.Event) []domain.Event {
	var out []domain.Event
	for _, b := range batches {
		out = append(out, b...)
	}
	return out
}
