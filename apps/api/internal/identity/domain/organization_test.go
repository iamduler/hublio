package domain_test

import (
	"testing"
	"time"

	"hublio/internal/identity/domain"

	"github.com/google/uuid"
)

func TestOrganizationLifecycle(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	id := uuid.MustParse("01900000-0000-7000-8000-000000000001")

	tests := []struct {
		name    string
		act     func(org *domain.Organization) error
		wantErr error
		status  domain.OrganizationStatus
		canIntent bool
	}{
		{
			name: "suspend from active",
			act: func(org *domain.Organization) error {
				return org.Suspend(now.Add(time.Minute))
			},
			status:    domain.OrganizationStatusSuspended,
			canIntent: false,
		},
		{
			name: "activate from suspended",
			act: func(org *domain.Organization) error {
				if err := org.Suspend(now.Add(time.Minute)); err != nil {
					return err
				}
				return org.Activate(now.Add(2 * time.Minute))
			},
			status:    domain.OrganizationStatusActive,
			canIntent: true,
		},
		{
			name: "cannot activate from active",
			act: func(org *domain.Organization) error {
				return org.Activate(now)
			},
			wantErr: domain.ErrInvalidTransition,
			status:  domain.OrganizationStatusActive,
			canIntent: true,
		},
		{
			name: "archive blocks intents",
			act: func(org *domain.Organization) error {
				return org.Archive(now.Add(time.Minute))
			},
			status:    domain.OrganizationStatusArchived,
			canIntent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := domain.NewOrganization(id, "Acme", now)
			if err != nil {
				t.Fatal(err)
			}
			err = tt.act(org)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("got err %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if org.Status() != tt.status {
				t.Fatalf("status %s want %s", org.Status(), tt.status)
			}
			if org.CanSubmitIntents() != tt.canIntent {
				t.Fatalf("CanSubmitIntents=%v want %v", org.CanSubmitIntents(), tt.canIntent)
			}
		})
	}
}

func TestOrganizationInvalidName(t *testing.T) {
	_, err := domain.NewOrganization(uuid.Nil, "  ", time.Now())
	if err != domain.ErrInvalidName {
		t.Fatalf("got %v", err)
	}
}
