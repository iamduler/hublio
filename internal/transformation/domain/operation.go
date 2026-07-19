package domain

// Operation applies one Canonical → Canonical transformation to a Document. Implementations
// must never depend on Provider DTOs, HTTP, persistence, or business rules (docs/06 §2-3):
// that keeps the Transformation Engine a pure, deterministic in-memory engine.
type Operation interface {
	Apply(doc Document) (Document, error)
}
