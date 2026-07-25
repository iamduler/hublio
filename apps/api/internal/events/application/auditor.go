package application

import (
	"context"
	"strings"

	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

// Auditor is the Events BC port other Bounded Contexts record audit facts through (via a
// bridge — see internal/events/infrastructure).
type Auditor interface {
	Record(ctx context.Context, rec AuditRecord) error
}

// AuditRecord is the Application-level shape of one audit_logs row.
type AuditRecord struct {
	OrganizationID *uuid.UUID
	WorkspaceID    *uuid.UUID
	ActorType      string // user|api_key|system
	ActorID        *uuid.UUID
	Action         string
	ResourceType   string
	ResourceID     *uuid.UUID
	RequestID      string
	CorrelationID  string
	IP             string
	UserAgent      string
	Metadata       map[string]any
}

// redactedMetadataKeys are never persisted, even if a caller mistakenly includes them
// (AGENTS.md: never log secrets). Comparison is case-insensitive substring match.
var redactedMetadataKeys = []string{"secret", "password", "plaintext", "token", "key_hash"}

// Record persists one AuditEntry (append-only, `audit_logs` table). Metadata is redacted
// before it ever reaches the Domain/Infrastructure layers.
func (s *Services) Record(ctx context.Context, rec AuditRecord) error {
	entryID, err := id.NewV7()
	if err != nil {
		return apperr.Wrap(err, "failed to generate audit entry id", apperr.ErrCodeInternal)
	}

	entry, err := domain.NewAuditEntry(
		entryID,
		rec.OrganizationID,
		rec.WorkspaceID,
		domain.ActorType(rec.ActorType),
		rec.ActorID,
		rec.Action,
		rec.ResourceType,
		rec.ResourceID,
		rec.RequestID,
		rec.CorrelationID,
		rec.IP,
		rec.UserAgent,
		redactMetadata(rec.Metadata),
		s.clock().Now(),
	)
	if err != nil {
		return apperr.Wrap(err, "invalid audit entry", apperr.ErrCodeBadRequest)
	}

	if err := s.auditRepo().Save(ctx, entry); err != nil {
		return apperr.Wrap(err, "failed to persist audit entry", apperr.ErrCodeInternal)
	}
	s.metricsOrNoop().IncAuditRecords()
	return nil
}

func (s *Services) auditRepo() domain.AuditRepository {
	if s.Audit != nil {
		return s.Audit
	}
	return noopAuditRepository{}
}

type noopAuditRepository struct{}

func (noopAuditRepository) Save(ctx context.Context, entry *domain.AuditEntry) error {
	_, _ = ctx, entry
	return nil
}

// redactMetadata drops any key that looks secret-ish, never persisting it (AGENTS.md: never
// log secrets/tokens/API keys/passwords).
func redactMetadata(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return metadata
	}
	out := make(map[string]any, len(metadata))
	for k, v := range metadata {
		if isRedactedKey(k) {
			continue
		}
		out[k] = v
	}
	return out
}

func isRedactedKey(key string) bool {
	lower := strings.ToLower(key)
	for _, redacted := range redactedMetadataKeys {
		if strings.Contains(lower, redacted) {
			return true
		}
	}
	return false
}
