package id

import "github.com/google/uuid"

// NewV7 generates an application-layer UUID v7. Repositories must never generate IDs.
func NewV7() (uuid.UUID, error) {
	return uuid.NewV7()
}

// MustV7 panics only in tests/bootstrap helpers; prefer NewV7 in production paths.
func MustV7() uuid.UUID {
	id, err := NewV7()
	if err != nil {
		panic(err)
	}
	return id
}
