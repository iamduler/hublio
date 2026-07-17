package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustConnector(t *testing.T) *Connector {
	t.Helper()
	c, err := NewConnector(uuid.Must(uuid.NewV7()), "fake", "Fake Connector", "hublio", ConnectorCategoryUtility, "1.0.0", "desc", "https://example.com", "https://example.com/docs", time.Now())
	if err != nil {
		t.Fatalf("NewConnector() unexpected error: %v", err)
	}
	return c
}

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		connNam string
		vendor  string
		cat     ConnectorCategory
		version string
		wantErr error
	}{
		{"valid", "nhanh", "Nhanh.vn", "Nhanh", ConnectorCategorySource, "1.0.0", nil},
		{"empty code", "", "Name", "Vendor", ConnectorCategorySource, "1.0.0", ErrInvalidCode},
		{"empty name", "code", "", "Vendor", ConnectorCategorySource, "1.0.0", ErrInvalidName},
		{"empty vendor", "code", "Name", "", ConnectorCategorySource, "1.0.0", ErrInvalidVendor},
		{"invalid category", "code", "Name", "Vendor", ConnectorCategory("bogus"), "1.0.0", ErrInvalidCategory},
		{"empty version", "code", "Name", "Vendor", ConnectorCategorySource, "", ErrInvalidVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConnector(uuid.Must(uuid.NewV7()), tt.code, tt.connNam, tt.vendor, tt.cat, tt.version, "", "", "", time.Now())
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnector_StatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState ConnectorStatus
		action    func(*Connector) error
		wantErr   error
		wantState ConnectorStatus
	}{
		{"registered enable succeeds", ConnectorStatusRegistered, func(c *Connector) error { return c.Enable(time.Now()) }, nil, ConnectorStatusEnabled},
		{"registered disable fails", ConnectorStatusRegistered, func(c *Connector) error { return c.Disable(time.Now()) }, ErrInvalidTransition, ConnectorStatusRegistered},
		{"enabled disable succeeds", ConnectorStatusEnabled, func(c *Connector) error { return c.Disable(time.Now()) }, nil, ConnectorStatusDisabled},
		{"enabled enable fails", ConnectorStatusEnabled, func(c *Connector) error { return c.Enable(time.Now()) }, ErrInvalidTransition, ConnectorStatusEnabled},
		{"disabled enable succeeds", ConnectorStatusDisabled, func(c *Connector) error { return c.Enable(time.Now()) }, nil, ConnectorStatusEnabled},
		{"enabled remove succeeds", ConnectorStatusEnabled, func(c *Connector) error { return c.Remove(time.Now()) }, nil, ConnectorStatusRemoved},
		{"removed enable fails", ConnectorStatusRemoved, func(c *Connector) error { return c.Enable(time.Now()) }, ErrConnectorRemoved, ConnectorStatusRemoved},
		{"removed remove fails", ConnectorStatusRemoved, func(c *Connector) error { return c.Remove(time.Now()) }, ErrInvalidTransition, ConnectorStatusRemoved},
		{"removed disable fails", ConnectorStatusRemoved, func(c *Connector) error { return c.Disable(time.Now()) }, ErrInvalidTransition, ConnectorStatusRemoved},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mustConnector(t)
			c.status = tt.fromState
			err := tt.action(c)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
			if c.Status() != tt.wantState {
				t.Fatalf("got status %v, want %v", c.Status(), tt.wantState)
			}
		})
	}
}

func TestConnector_IsUsable(t *testing.T) {
	c := mustConnector(t)
	if c.IsUsable() {
		t.Fatalf("registered connector should not be usable")
	}
	if err := c.Enable(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsUsable() {
		t.Fatalf("enabled connector should be usable")
	}
	if err := c.Disable(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.IsUsable() {
		t.Fatalf("disabled connector should not be usable")
	}
}

func TestConnector_AddCapability(t *testing.T) {
	c := mustConnector(t)
	capID := uuid.Must(uuid.NewV7())
	cap, err := c.AddCapability(capID, "create_invoice", "Create Invoice", true, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Code() != "create_invoice" || !cap.IsAsync() || !cap.IsEnabled() {
		t.Fatalf("unexpected capability state: %+v", cap)
	}
	if len(c.Capabilities()) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(c.Capabilities()))
	}

	if _, err := c.AddCapability(uuid.Must(uuid.NewV7()), "create_invoice", "Dup", false, time.Now()); err != ErrConflict {
		t.Fatalf("expected conflict on duplicate capability code, got %v", err)
	}
}

func TestConnector_AddCapability_RemovedConnector(t *testing.T) {
	c := mustConnector(t)
	if err := c.Enable(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Remove(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := c.AddCapability(uuid.Must(uuid.NewV7()), "x", "X", false, time.Now()); err != ErrConnectorRemoved {
		t.Fatalf("expected ErrConnectorRemoved, got %v", err)
	}
}

func TestCapability_EnableDisable(t *testing.T) {
	c := mustConnector(t)
	cap, err := c.AddCapability(uuid.Must(uuid.NewV7()), "sync", "Sync", false, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cap.Enable(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition enabling already-enabled capability, got %v", err)
	}
	if err := cap.Disable(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.IsEnabled() {
		t.Fatalf("expected capability disabled")
	}
	if err := cap.Disable(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition disabling already-disabled capability, got %v", err)
	}
}

func TestConnector_PullEvents(t *testing.T) {
	c := mustConnector(t)
	events := c.PullEvents()
	if len(events) != 1 || events[0].Name != EventConnectorRegistered {
		t.Fatalf("expected 1 ConnectorRegistered event, got %+v", events)
	}
	if len(c.PullEvents()) != 0 {
		t.Fatalf("expected events drained after pull")
	}
	_ = c.Enable(time.Now())
	events = c.PullEvents()
	if len(events) != 1 || events[0].Name != EventConnectorEnabled {
		t.Fatalf("expected 1 ConnectorEnabled event, got %+v", events)
	}
}
