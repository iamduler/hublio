package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"

	"hublio/internal/integration/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("integration repo: %w", err)
}

func mapUnique(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return fmt.Errorf("integration repo: %w", err)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strFromPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgtypeToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func marshalJSONMap(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func unmarshalJSONMap(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// encryptedSecretEnvelope is the JSONB shape stored in credentials.encrypted_secret.
// The Domain treats the whole thing as opaque ciphertext; only Infrastructure knows the envelope shape.
type encryptedSecretEnvelope struct {
	Ciphertext string `json:"ciphertext"`
}

func marshalEncryptedSecret(ciphertext []byte) ([]byte, error) {
	return json.Marshal(encryptedSecretEnvelope{Ciphertext: string(ciphertext)})
}

func unmarshalEncryptedSecret(data []byte) ([]byte, error) {
	var env encryptedSecretEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return []byte(env.Ciphertext), nil
}
