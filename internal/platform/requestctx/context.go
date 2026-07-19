package requestctx

import "context"

type contextKey string

const (
	KeyTraceID        contextKey = "trace_id"
	KeyCorrelationID  contextKey = "correlation_id"
	KeyRequestID      contextKey = "request_id"
	KeyOrganizationID contextKey = "organization_id"
	KeyWorkspaceID    contextKey = "workspace_id"
	KeyUserID         contextKey = "user_id"
	KeyAPIKeyID       contextKey = "api_key_id"
	KeyIP             contextKey = "ip"
	KeyUserAgent      contextKey = "user_agent"
)

func With(ctx context.Context, key contextKey, value string) context.Context {
	if value == "" {
		return ctx
	}
	return context.WithValue(ctx, key, value)
}

func Get(ctx context.Context, key contextKey) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

func TraceID(ctx context.Context) string        { return Get(ctx, KeyTraceID) }
func CorrelationID(ctx context.Context) string  { return Get(ctx, KeyCorrelationID) }
func RequestID(ctx context.Context) string      { return Get(ctx, KeyRequestID) }
func OrganizationID(ctx context.Context) string { return Get(ctx, KeyOrganizationID) }
func WorkspaceID(ctx context.Context) string    { return Get(ctx, KeyWorkspaceID) }
func UserID(ctx context.Context) string         { return Get(ctx, KeyUserID) }
func APIKeyID(ctx context.Context) string       { return Get(ctx, KeyAPIKeyID) }
func IP(ctx context.Context) string             { return Get(ctx, KeyIP) }
func UserAgent(ctx context.Context) string      { return Get(ctx, KeyUserAgent) }
