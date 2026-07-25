package infrastructure

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func timeFrom(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time.UTC()
}

func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
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

func ipPtr(s string) *netip.Addr {
	if s == "" {
		return nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return nil
	}
	return &addr
}

func ipFromPtr(p *netip.Addr) string {
	if p == nil || !p.IsValid() {
		return ""
	}
	return p.String()
}

func marshalJSONMap(m map[string]any) ([]byte, error) {
	if len(m) == 0 {
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
