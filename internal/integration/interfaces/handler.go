package interfaces

import (
	"context"
	"net/http"
	"time"

	"hublio/internal/integration/application"
	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MembershipChecker verifies that a User belongs to a Workspace. Implemented by a thin
// adapter over the Identity BC in the composition root (internal/platform/server).
type MembershipChecker interface {
	Check(ctx context.Context, workspaceID, userID uuid.UUID) error
}

type Handler struct {
	svc        *application.Services
	pool       *pgxpool.Pool
	membership MembershipChecker
}

func NewHandler(svc *application.Services, pool *pgxpool.Pool, membership MembershipChecker) *Handler {
	return &Handler{svc: svc, pool: pool, membership: membership}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, jwtAuth gin.HandlerFunc) {
	integration := api.Group("/integration")
	integration.Use(jwtAuth)
	{
		integration.GET("/connectors", h.listConnectors)
		integration.POST("/connectors", h.registerConnector)
		integration.GET("/connectors/:connectorId", h.getConnector)
		integration.POST("/connectors/:connectorId/enable", h.enableConnector)
		integration.POST("/connectors/:connectorId/disable", h.disableConnector)
		integration.POST("/connectors/:connectorId/remove", h.removeConnector)

		integration.GET("/workspaces/:workspaceId/connections", h.listConnections)
		integration.POST("/workspaces/:workspaceId/connections", h.createConnection)
		integration.GET("/workspaces/:workspaceId/connections/:connectionId", h.getConnection)
		integration.POST("/workspaces/:workspaceId/connections/:connectionId/verify", h.verifyConnection)
		integration.POST("/workspaces/:workspaceId/connections/:connectionId/disable", h.disableConnection)
		integration.POST("/workspaces/:workspaceId/connections/:connectionId/enable", h.enableConnection)
		integration.POST("/workspaces/:workspaceId/connections/:connectionId/credentials/rotate", h.rotateCredential)
	}
}

type capabilityRequest struct {
	Code        string `json:"code" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	IsAsync     bool   `json:"is_async"`
}

type registerConnectorRequest struct {
	Code             string              `json:"code" binding:"required"`
	Name             string              `json:"name" binding:"required"`
	Vendor           string              `json:"vendor" binding:"required"`
	Category         string              `json:"category" binding:"required"`
	Version          string              `json:"version" binding:"required"`
	Description      string              `json:"description"`
	Homepage         string              `json:"homepage"`
	DocumentationURL string              `json:"documentation_url"`
	Capabilities     []capabilityRequest `json:"capabilities"`
}

func (h *Handler) registerConnector(c *gin.Context) {
	if _, ok := actorUserID(c); !ok {
		return
	}
	var req registerConnectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	category, err := domain.ParseConnectorCategory(req.Category)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid category", apperr.ErrCodeBadRequest))
		return
	}
	caps := make([]application.RegisterCapabilityInput, 0, len(req.Capabilities))
	for _, capReq := range req.Capabilities {
		caps = append(caps, application.RegisterCapabilityInput{
			Code:        capReq.Code,
			DisplayName: capReq.DisplayName,
			IsAsync:     capReq.IsAsync,
		})
	}

	var connector *domain.Connector
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		connector, innerErr = h.svc.RegisterConnector(ctx, application.RegisterConnectorInput{
			Code:             req.Code,
			Name:             req.Name,
			Vendor:           req.Vendor,
			Category:         category,
			Version:          req.Version,
			Description:      req.Description,
			Homepage:         req.Homepage,
			DocumentationURL: req.DocumentationURL,
			Capabilities:     caps,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), connector.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusCreated, "connector registered", connectorDTO(connector))
}

func (h *Handler) listConnectors(c *gin.Context) {
	if _, ok := actorUserID(c); !ok {
		return
	}
	list, err := h.svc.ListConnectors(c.Request.Context())
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, connector := range list {
		out = append(out, connectorDTO(connector))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "connectors", out)
}

func (h *Handler) getConnector(c *gin.Context) {
	if _, ok := actorUserID(c); !ok {
		return
	}
	connectorID, ok := parseUUIDParam(c, "connectorId")
	if !ok {
		return
	}
	connector, err := h.svc.GetConnector(c.Request.Context(), connectorID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "connector", connectorDTO(connector))
}

func (h *Handler) enableConnector(c *gin.Context) {
	h.connectorLifecycle(c, h.svc.EnableConnector, "connector enabled")
}

func (h *Handler) disableConnector(c *gin.Context) {
	h.connectorLifecycle(c, h.svc.DisableConnector, "connector disabled")
}

func (h *Handler) removeConnector(c *gin.Context) {
	h.connectorLifecycle(c, h.svc.RemoveConnector, "connector removed")
}

func (h *Handler) connectorLifecycle(c *gin.Context, fn func(context.Context, uuid.UUID) (*domain.Connector, error), message string) {
	if _, ok := actorUserID(c); !ok {
		return
	}
	connectorID, ok := parseUUIDParam(c, "connectorId")
	if !ok {
		return
	}
	var connector *domain.Connector
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		connector, innerErr = fn(ctx, connectorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), connector.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, message, connectorDTO(connector))
}

type createConnectionRequest struct {
	ConnectorID         string         `json:"connector_id" binding:"required"`
	Name                string         `json:"name" binding:"required"`
	IsDefault           bool           `json:"is_default"`
	Description         string         `json:"description"`
	Environment         string         `json:"environment" binding:"required"`
	Config              map[string]any `json:"config"`
	RetryPolicy         map[string]any `json:"retry_policy"`
	TimeoutSeconds      int            `json:"timeout_seconds"`
	CredentialType      string         `json:"credential_type" binding:"required"`
	Secret              map[string]any `json:"secret" binding:"required"`
	CredentialExpiresAt *time.Time     `json:"credential_expires_at"`
}

func (h *Handler) createConnection(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := h.requireWorkspaceMember(c, workspaceID)
	if !ok {
		return
	}
	var req createConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	connectorID, err := uuid.Parse(req.ConnectorID)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid connector_id", apperr.ErrCodeBadRequest))
		return
	}
	credType, err := domain.ParseCredentialType(req.CredentialType)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid credential_type", apperr.ErrCodeBadRequest))
		return
	}

	var result *application.CreateConnectionResult
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.CreateConnection(ctx, application.CreateConnectionInput{
			WorkspaceID:         workspaceID,
			ConnectorID:         connectorID,
			Name:                req.Name,
			IsDefault:           req.IsDefault,
			Description:         req.Description,
			Environment:         req.Environment,
			Config:              req.Config,
			RetryPolicy:         req.RetryPolicy,
			TimeoutSeconds:      req.TimeoutSeconds,
			CredentialType:      credType,
			Secret:              req.Secret,
			CredentialExpiresAt: req.CredentialExpiresAt,
			ActorUserID:         actorID,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(),
		appendMany(result.Connection.PullEvents(), result.Credential.PullEvents())...,
	)
	httpx.ResponseSuccess(c, http.StatusCreated, "connection created", gin.H{
		"connection": connectionDTO(result.Connection),
		"credential": credentialDTO(result.Credential),
	})
}

func (h *Handler) listConnections(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	list, err := h.svc.ListConnections(c.Request.Context(), workspaceID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, conn := range list {
		out = append(out, connectionDTO(conn))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "connections", out)
}

func (h *Handler) getConnection(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	connectionID, ok := parseUUIDParam(c, "connectionId")
	if !ok {
		return
	}
	conn, err := h.svc.GetConnection(c.Request.Context(), workspaceID, connectionID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "connection", connectionDTO(conn))
}

func (h *Handler) verifyConnection(c *gin.Context) {
	h.connectionLifecycle(c, func(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
		return h.svc.VerifyConnection(ctx, workspaceID, connectionID)
	}, "connection verification finished")
}

func (h *Handler) disableConnection(c *gin.Context) {
	h.connectionLifecycle(c, h.svc.DisableConnection, "connection disabled")
}

func (h *Handler) enableConnection(c *gin.Context) {
	h.connectionLifecycle(c, h.svc.EnableConnection, "connection enabled")
}

func (h *Handler) connectionLifecycle(
	c *gin.Context,
	fn func(context.Context, uuid.UUID, uuid.UUID) (*domain.Connection, error),
	message string,
) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	connectionID, ok := parseUUIDParam(c, "connectionId")
	if !ok {
		return
	}
	var conn *domain.Connection
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		conn, innerErr = fn(ctx, workspaceID, connectionID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), conn.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, message, connectionDTO(conn))
}

type rotateCredentialRequest struct {
	CredentialType string         `json:"credential_type" binding:"required"`
	Secret         map[string]any `json:"secret" binding:"required"`
	ExpiresAt      *time.Time     `json:"expires_at"`
}

func (h *Handler) rotateCredential(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := h.requireWorkspaceMember(c, workspaceID)
	if !ok {
		return
	}
	connectionID, ok := parseUUIDParam(c, "connectionId")
	if !ok {
		return
	}
	var req rotateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	credType, err := domain.ParseCredentialType(req.CredentialType)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid credential_type", apperr.ErrCodeBadRequest))
		return
	}

	var cred *domain.Credential
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		cred, innerErr = h.svc.RotateCredential(ctx, application.RotateCredentialInput{
			WorkspaceID:    workspaceID,
			ConnectionID:   connectionID,
			CredentialType: credType,
			Secret:         req.Secret,
			ExpiresAt:      req.ExpiresAt,
			ActorUserID:    actorID,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), cred.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusCreated, "credential rotated", credentialDTO(cred))
}

func (h *Handler) requireWorkspaceMember(c *gin.Context, workspaceID uuid.UUID) (uuid.UUID, bool) {
	actorID, ok := actorUserID(c)
	if !ok {
		return uuid.Nil, false
	}
	if h.membership != nil {
		if err := h.membership.Check(c.Request.Context(), workspaceID, actorID); err != nil {
			httpx.ResponseError(c, err)
			return uuid.Nil, false
		}
	}
	return actorID, true
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

func connectorDTO(connector *domain.Connector) gin.H {
	caps := make([]gin.H, 0, len(connector.Capabilities()))
	for _, capability := range connector.Capabilities() {
		caps = append(caps, capabilityDTO(capability))
	}
	return gin.H{
		"id":                connector.ID().String(),
		"code":              connector.Code(),
		"name":              connector.Name(),
		"vendor":            connector.Vendor(),
		"category":          string(connector.Category()),
		"version":           connector.Version(),
		"status":            string(connector.Status()),
		"description":       connector.Description(),
		"homepage":          connector.Homepage(),
		"documentation_url": connector.DocumentationURL(),
		"capabilities":      caps,
		"created_at":        connector.CreatedAt(),
		"updated_at":        connector.UpdatedAt(),
	}
}

func capabilityDTO(capability *domain.Capability) gin.H {
	return gin.H{
		"id":           capability.ID().String(),
		"code":         capability.Code(),
		"display_name": capability.DisplayName(),
		"status":       string(capability.Status()),
		"is_async":     capability.IsAsync(),
	}
}

func connectionDTO(conn *domain.Connection) gin.H {
	var activeCredentialID any
	if id := conn.ActiveCredentialID(); id != nil {
		activeCredentialID = id.String()
	}
	return gin.H{
		"id":                   conn.ID().String(),
		"workspace_id":         conn.WorkspaceID().String(),
		"connector_id":         conn.ConnectorID().String(),
		"name":                 conn.Name(),
		"is_default":           conn.IsDefault(),
		"description":          conn.Description(),
		"environment":          conn.Environment(),
		"status":               string(conn.Status()),
		"config":               conn.Config(),
		"retry_policy":         conn.RetryPolicy(),
		"timeout_seconds":      conn.TimeoutSeconds(),
		"active_credential_id": activeCredentialID,
		"created_at":           conn.CreatedAt(),
		"updated_at":           conn.UpdatedAt(),
	}
}

// credentialDTO never includes the encrypted or plaintext secret.
func credentialDTO(cred *domain.Credential) gin.H {
	return gin.H{
		"id":            cred.ID().String(),
		"connection_id": cred.ConnectionID().String(),
		"type":          string(cred.Type()),
		"status":        string(cred.Status()),
		"version":       cred.Version(),
		"expires_at":    cred.ExpiresAt(),
		"rotated_at":    cred.RotatedAt(),
		"created_at":    cred.CreatedAt(),
		"updated_at":    cred.UpdatedAt(),
	}
}

func appendMany(batches ...[]domain.Event) []domain.Event {
	var out []domain.Event
	for _, b := range batches {
		out = append(out, b...)
	}
	return out
}
