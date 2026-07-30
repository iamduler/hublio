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
	"hublio/internal/platform/requestctx"

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
		authGroup.POST("/forgot-password", h.forgotPassword)
		authGroup.POST("/reset-password", h.resetPassword)
		authGroup.POST("/verify-email/request", h.requestEmailVerification)
		authGroup.POST("/verify-email", h.verifyEmail)
		authGroup.GET("/oauth/providers", h.oauthProviders)
		authGroup.POST("/oauth/callback", h.oauthCallback)
		authGroup.GET("/oauth/onboarding", h.oauthOnboardingPreview)
		authGroup.POST("/oauth/complete-registration", h.oauthCompleteRegistration)
		authGroup.POST("/mfa/verify", h.mfaVerify)
	}

	mfaGroup := api.Group("/auth/mfa")
	mfaGroup.Use(jwtAuth)
	{
		mfaGroup.GET("/status", h.mfaStatus)
		mfaGroup.POST("/setup", h.mfaSetup)
		mfaGroup.POST("/enable", h.mfaEnable)
		mfaGroup.POST("/disable", h.mfaDisable)
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
		identity.GET("/workspaces/:workspaceId/members", h.listMembers)
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

	// Best-effort: send the email verification code after a successful registration. Run it
	// detached from the request so mail latency never delays the response, and never fail
	// registration if mail (or the code store) is unavailable; failures are logged downstream.
	email := result.User.Email()
	go func() {
		_ = h.svc.RequestEmailVerification(context.Background(), email)
	}()

	httpx.ResponseSuccess(c, http.StatusCreated, "registered", gin.H{
		"organization": organizationDTO(result.Organization),
		"workspace":    workspaceDTO(result.Workspace),
		"user":         userDTO(result.User),
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"`
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
			DeviceID: req.DeviceID,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	if result.MFARequired {
		h.recordLoginAudit(c, result.User, "user.mfa_challenged")
		httpx.ResponseSuccess(c, http.StatusOK, "mfa required", gin.H{
			"mfa_required": true,
			"mfa_token":    result.MFAToken,
		})
		return
	}

	h.recordLoginAudit(c, result.User, "user.login")
	httpx.ResponseSuccess(c, http.StatusOK, "logged in", loginDTO(result))
}

// recordLoginAudit stamps the user's organization onto the audit context, which is not yet
// available from the request (login happens before any JWT/workspace scoping).
func (h *Handler) recordLoginAudit(c *gin.Context, user *domain.User, action string) {
	auditCtx := requestctx.With(c.Request.Context(), requestctx.KeyOrganizationID, user.OrganizationID().String())
	h.svc.RecordAudit(auditCtx, application.AuditEvent{
		ActorType:    "user",
		ActorID:      user.ID(),
		Action:       action,
		ResourceType: "user",
		ResourceID:   user.ID(),
	})
}

type mfaVerifyRequest struct {
	MFAToken     string `json:"mfa_token" binding:"required"`
	Code         string `json:"code"`
	RecoveryCode string `json:"recovery_code"`
	DeviceID     string `json:"device_id"`
	TrustDevice  bool   `json:"trust_device"`
}

func (h *Handler) mfaVerify(c *gin.Context) {
	var req mfaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var result *application.LoginResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.VerifyMFA(ctx, h.tokens, application.VerifyMFAInput{
			Token:        req.MFAToken,
			Code:         req.Code,
			RecoveryCode: req.RecoveryCode,
			DeviceID:     req.DeviceID,
			TrustDevice:  req.TrustDevice,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	h.recordLoginAudit(c, result.User, "user.login")
	httpx.ResponseSuccess(c, http.StatusOK, "logged in", loginDTO(result))
}

func (h *Handler) mfaStatus(c *gin.Context) {
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var status *application.MFAStatus
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		status, innerErr = h.svc.GetMFAStatus(ctx, actorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "mfa status", gin.H{
		"enabled":                   status.Enabled,
		"pending_enrollment":        status.PendingEnrollment,
		"remaining_recovery_codes":  status.RemainingRecoveryCodes,
		"can_enroll":                status.CanEnroll,
	})
}

func (h *Handler) mfaSetup(c *gin.Context) {
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var result *application.MFASetupResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.SetupMFA(ctx, actorID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "user.mfa_setup",
		ResourceType: "user",
		ResourceID:   actorID,
	})
	httpx.ResponseSuccess(c, http.StatusOK, "mfa enrollment started", gin.H{
		"secret":         result.Secret,
		"otpauth_url":    result.OTPAuthURL,
		"recovery_codes": result.RecoveryCodes,
		"warning":        "store the recovery codes now; they will not be shown again",
	})
}

type mfaEnableRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

func (h *Handler) mfaEnable(c *gin.Context) {
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var req mfaEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		return h.svc.EnableMFA(ctx, actorID, req.Code)
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "user.mfa_enabled",
		ResourceType: "user",
		ResourceID:   actorID,
	})
	httpx.ResponseSuccess(c, http.StatusOK, "mfa enabled", nil)
}

type mfaDisableRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *Handler) mfaDisable(c *gin.Context) {
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	var req mfaDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		return h.svc.DisableMFA(ctx, actorID, req.Password)
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "user.mfa_disabled",
		ResourceType: "user",
		ResourceID:   actorID,
	})
	httpx.ResponseSuccess(c, http.StatusOK, "mfa disabled", nil)
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

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *Handler) forgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	// Anti-enumeration: RequestPasswordReset never reveals whether the email exists.
	if err := h.svc.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "if an account exists for that email, a reset link has been sent", nil)
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) resetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		return h.svc.ResetPassword(ctx, application.ResetPasswordInput{
			Token:    req.Token,
			Password: req.Password,
		})
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "password updated", nil)
}

type requestEmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *Handler) requestEmailVerification(c *gin.Context) {
	var req requestEmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	// Anti-enumeration: RequestEmailVerification never reveals whether the email exists.
	if err := h.svc.RequestEmailVerification(c.Request.Context(), req.Email); err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "if an account exists for that email, a verification code has been sent", nil)
}

type verifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

func (h *Handler) verifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		return h.svc.VerifyEmail(ctx, req.Email, req.Code)
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "email verified", nil)
}

type oauthCallbackRequest struct {
	Provider     string `json:"provider" binding:"required"`
	Code         string `json:"code" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
}

func (h *Handler) oauthProviders(c *gin.Context) {
	providers := []string{}
	if h.svc.OAuth != nil {
		for _, p := range []domain.OAuthProvider{
			domain.OAuthProviderGoogle,
			domain.OAuthProviderMicrosoft,
			domain.OAuthProviderGitHub,
		} {
			if h.svc.OAuth.Configured(p) {
				providers = append(providers, string(p))
			}
		}
	}
	httpx.ResponseSuccess(c, http.StatusOK, "oauth providers", gin.H{
		"providers": providers,
	})
}

func (h *Handler) oauthCallback(c *gin.Context) {
	var req oauthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var result *application.OAuthCallbackResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.OAuthCallback(ctx, h.tokens, application.OAuthCallbackInput{
			Provider:     req.Provider,
			Code:         req.Code,
			CodeVerifier: req.CodeVerifier,
			RedirectURI:  req.RedirectURI,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	if result.Status == "onboarding_required" {
		httpx.ResponseSuccess(c, http.StatusOK, "onboarding required", gin.H{
			"status":           result.Status,
			"onboarding_token": result.OnboardingToken,
			"email":            result.Email,
			"full_name":        result.FullName,
		})
		return
	}

	auditCtx := requestctx.With(c.Request.Context(), requestctx.KeyOrganizationID, result.User.OrganizationID().String())
	h.svc.RecordAudit(auditCtx, application.AuditEvent{
		ActorType:    "user",
		ActorID:      result.User.ID(),
		Action:       "user.oauth_login",
		ResourceType: "user",
		ResourceID:   result.User.ID(),
	})

	httpx.ResponseSuccess(c, http.StatusOK, "logged in", gin.H{
		"status":        result.Status,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          userDTO(result.User),
	})
}

func (h *Handler) oauthOnboardingPreview(c *gin.Context) {
	token := strings.TrimSpace(c.GetHeader("X-OAuth-Onboarding-Token"))
	if token == "" {
		token = strings.TrimSpace(c.Query("token"))
	}
	preview, err := h.svc.PeekOAuthOnboarding(token)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "onboarding preview", gin.H{
		"email":     preview.Email,
		"full_name": preview.FullName,
		"provider":  preview.Provider,
	})
}

type oauthCompleteRequest struct {
	OnboardingToken  string `json:"onboarding_token" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
	WorkspaceName    string `json:"workspace_name"`
	Environment      string `json:"environment"`
}

func (h *Handler) oauthCompleteRegistration(c *gin.Context) {
	var req oauthCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var (
		login    *application.LoginResult
		org      *domain.Organization
		ws       *domain.Workspace
		mem      *domain.Membership
		identity *domain.OAuthIdentity
	)
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		login, org, ws, mem, identity, innerErr = h.svc.CompleteOAuthRegistration(ctx, h.tokens, application.CompleteOAuthRegistrationInput{
			OnboardingToken:  req.OnboardingToken,
			OrganizationName: req.OrganizationName,
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
			org.PullEvents(),
			login.User.PullEvents(),
			ws.PullEvents(),
			mem.PullEvents(),
		)...,
	)
	_ = identity

	auditCtx := requestctx.With(c.Request.Context(), requestctx.KeyOrganizationID, login.User.OrganizationID().String())
	h.svc.RecordAudit(auditCtx, application.AuditEvent{
		ActorType:    "user",
		ActorID:      login.User.ID(),
		Action:       "user.oauth_register",
		ResourceType: "user",
		ResourceID:   login.User.ID(),
	})

	httpx.ResponseSuccess(c, http.StatusCreated, "registered", gin.H{
		"access_token":  login.AccessToken,
		"refresh_token": login.RefreshToken,
		"organization":  organizationDTO(org),
		"workspace":     workspaceDTO(ws),
		"user":          userDTO(login.User),
	})
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

func (h *Handler) listMembers(c *gin.Context) {
	wsID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := actorUserID(c)
	if !ok {
		return
	}
	members, err := h.svc.ListWorkspaceMembers(c.Request.Context(), wsID, actorID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(members))
	for _, m := range members {
		out = append(out, gin.H{
			"user_id":    m.UserID().String(),
			"email":      m.Email(),
			"full_name":  m.FullName(),
			"role":       string(m.Role()),
			"created_at": m.CreatedAt(),
		})
	}
	httpx.ResponseSuccess(c, http.StatusOK, "members", out)
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
	orgID, _ := uuid.Parse(requestctx.OrganizationID(c.Request.Context()))
	events := application.EnrichEvents(result.APIKey.PullEvents(), orgID, wsID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "api_key.create",
		ResourceType: "api_key",
		ResourceID:   result.APIKey.ID(),
		Metadata:     map[string]any{"name": req.Name, "workspace_id": wsID.String()},
	})
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
	orgID, _ := uuid.Parse(requestctx.OrganizationID(c.Request.Context()))
	events := application.EnrichEvents(key.PullEvents(), orgID, wsID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "api_key.disable",
		ResourceType: "api_key",
		ResourceID:   keyID,
		Metadata:     map[string]any{"workspace_id": wsID.String()},
	})
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
	orgID, _ := uuid.Parse(requestctx.OrganizationID(c.Request.Context()))
	events := application.EnrichEvents(result.APIKey.PullEvents(), orgID, wsID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "api_key.rotate",
		ResourceType: "api_key",
		ResourceID:   result.APIKey.ID(),
		Metadata:     map[string]any{"workspace_id": wsID.String()},
	})
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

// loginDTO is the completed-login payload shared by password login and MFA verification.
func loginDTO(result *application.LoginResult) gin.H {
	return gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          userDTO(result.User),
	}
}

func userDTO(u *domain.User) gin.H {
	return gin.H{
		"id":                u.ID().String(),
		"organization_id":   u.OrganizationID().String(),
		"email":             u.Email(),
		"full_name":         u.FullName(),
		"status":            string(u.Status()),
		"is_platform_admin": u.IsPlatformAdmin(),
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
