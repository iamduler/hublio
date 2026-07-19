package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type SyncRouteStatus string

const (
	SyncRouteStatusDraft    SyncRouteStatus = "draft"
	SyncRouteStatusEnabled  SyncRouteStatus = "enabled"
	SyncRouteStatusDisabled SyncRouteStatus = "disabled"
)

type SyncRouteTrigger string

const (
	SyncRouteTriggerWebhook  SyncRouteTrigger = "webhook"
	SyncRouteTriggerSchedule SyncRouteTrigger = "schedule"
	SyncRouteTriggerBoth     SyncRouteTrigger = "both"
)

func ParseSyncRouteTrigger(v string) (SyncRouteTrigger, error) {
	switch SyncRouteTrigger(strings.TrimSpace(v)) {
	case SyncRouteTriggerWebhook, SyncRouteTriggerSchedule, SyncRouteTriggerBoth:
		return SyncRouteTrigger(strings.TrimSpace(v)), nil
	default:
		return "", ErrInvalidSyncRouteTrigger
	}
}

type ActivityGroupMode string

const (
	ActivityGroupSequential ActivityGroupMode = "sequential"
	ActivityGroupParallel   ActivityGroupMode = "parallel"
)

func ParseActivityGroupMode(v string) (ActivityGroupMode, error) {
	switch ActivityGroupMode(strings.TrimSpace(v)) {
	case ActivityGroupSequential, ActivityGroupParallel:
		return ActivityGroupMode(strings.TrimSpace(v)), nil
	default:
		return "", ErrInvalidActivityGroup
	}
}

// ActivityStep is one destination action under a SyncRoute activity group.
// Runtime maps each step to one Execution under the accepting Intent (fan-out).
type ActivityStep struct {
	DestinationConnectionID uuid.UUID
	Capability              string
	MappingKey              string
}

// ActivityGroup is an ordered fan-out group: sequential or parallel Executions.
type ActivityGroup struct {
	Mode  ActivityGroupMode
	Steps []ActivityStep
}

// ReverseConfig optionally updates the source Connection after primary activities.
type ReverseConfig struct {
	Capability string
	On         string // success | failure | always
}

func ParseReverseOn(v string) (string, error) {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "", "success":
		return "success", nil
	case "failure", "always":
		return v, nil
	default:
		return "", ErrInvalidReverseConfig
	}
}

// SyncRoute is Workspace-scoped Integration configuration (origin → destinations).
// It is not a Runtime Aggregate and not a Workflow — Orchestration owns Intent/Execution.
type SyncRoute struct {
	eventRecorder

	id                  uuid.UUID
	workspaceID         uuid.UUID
	sourceConnectionID  uuid.UUID
	name                string
	status              SyncRouteStatus
	trigger             SyncRouteTrigger
	resourceTypes       []string
	schedule            map[string]any
	filter              map[string]any
	idempotencyRule     map[string]any
	activities          []ActivityGroup
	reverse             *ReverseConfig
	retryPolicy         map[string]any
	webhookSecretCipher []byte
	createdAt           time.Time
	updatedAt           time.Time
	deletedAt           *time.Time
}

type NewSyncRouteParams struct {
	ID                 uuid.UUID
	WorkspaceID        uuid.UUID
	SourceConnectionID uuid.UUID
	Name               string
	Trigger            SyncRouteTrigger
	ResourceTypes      []string
	Schedule           map[string]any
	Filter             map[string]any
	IdempotencyRule    map[string]any
	Activities         []ActivityGroup
	Reverse            *ReverseConfig
	RetryPolicy        map[string]any
	Now                time.Time
}

func NewSyncRoute(p NewSyncRouteParams) (*SyncRoute, error) {
	name := strings.TrimSpace(p.Name)
	if p.ID == uuid.Nil || p.WorkspaceID == uuid.Nil || p.SourceConnectionID == uuid.Nil {
		return nil, ErrInvalidName
	}
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}
	if _, err := ParseSyncRouteTrigger(string(p.Trigger)); err != nil {
		return nil, err
	}
	resourceTypes, err := normalizeResourceTypes(p.ResourceTypes)
	if err != nil {
		return nil, err
	}
	activities, err := normalizeActivities(p.Activities)
	if err != nil {
		return nil, err
	}
	reverse, err := normalizeReverse(p.Reverse)
	if err != nil {
		return nil, err
	}

	route := &SyncRoute{
		id:                 p.ID,
		workspaceID:        p.WorkspaceID,
		sourceConnectionID: p.SourceConnectionID,
		name:               name,
		status:             SyncRouteStatusDraft,
		trigger:            p.Trigger,
		resourceTypes:      resourceTypes,
		schedule:           p.Schedule,
		filter:             p.Filter,
		idempotencyRule:    p.IdempotencyRule,
		activities:         activities,
		reverse:            reverse,
		retryPolicy:        p.RetryPolicy,
		createdAt:          p.Now.UTC(),
		updatedAt:          p.Now.UTC(),
	}
	route.record(EventSyncRouteCreated, p.ID, p.Now.UTC(), map[string]any{
		"workspace_id":         p.WorkspaceID.String(),
		"source_connection_id": p.SourceConnectionID.String(),
		"name":                 name,
		"trigger_type":         string(p.Trigger),
	})
	return route, nil
}

func ReconstituteSyncRoute(
	id, workspaceID, sourceConnectionID uuid.UUID,
	name string,
	status SyncRouteStatus,
	trigger SyncRouteTrigger,
	resourceTypes []string,
	schedule, filter, idempotencyRule map[string]any,
	activities []ActivityGroup,
	reverse *ReverseConfig,
	retryPolicy map[string]any,
	webhookSecretCipher []byte,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *SyncRoute {
	return &SyncRoute{
		id:                  id,
		workspaceID:         workspaceID,
		sourceConnectionID:  sourceConnectionID,
		name:                name,
		status:              status,
		trigger:             trigger,
		resourceTypes:       resourceTypes,
		schedule:            schedule,
		filter:              filter,
		idempotencyRule:     idempotencyRule,
		activities:          activities,
		reverse:             reverse,
		retryPolicy:         retryPolicy,
		webhookSecretCipher: append([]byte(nil), webhookSecretCipher...),
		createdAt:           createdAt,
		updatedAt:           updatedAt,
		deletedAt:           deletedAt,
	}
}

func (r *SyncRoute) ID() uuid.UUID                 { return r.id }
func (r *SyncRoute) WorkspaceID() uuid.UUID        { return r.workspaceID }
func (r *SyncRoute) SourceConnectionID() uuid.UUID { return r.sourceConnectionID }
func (r *SyncRoute) Name() string                  { return r.name }
func (r *SyncRoute) Status() SyncRouteStatus       { return r.status }
func (r *SyncRoute) Trigger() SyncRouteTrigger     { return r.trigger }
func (r *SyncRoute) ResourceTypes() []string       { return append([]string(nil), r.resourceTypes...) }
func (r *SyncRoute) Schedule() map[string]any      { return r.schedule }
func (r *SyncRoute) Filter() map[string]any        { return r.filter }
func (r *SyncRoute) IdempotencyRule() map[string]any {
	return r.idempotencyRule
}
func (r *SyncRoute) Activities() []ActivityGroup {
	return append([]ActivityGroup(nil), r.activities...)
}
func (r *SyncRoute) Reverse() *ReverseConfig {
	if r.reverse == nil {
		return nil
	}
	cp := *r.reverse
	return &cp
}
func (r *SyncRoute) RetryPolicy() map[string]any { return r.retryPolicy }
func (r *SyncRoute) WebhookSecretCiphertext() []byte {
	return append([]byte(nil), r.webhookSecretCipher...)
}
func (r *SyncRoute) HasWebhookSecret() bool { return len(r.webhookSecretCipher) > 0 }
func (r *SyncRoute) CreatedAt() time.Time   { return r.createdAt }
func (r *SyncRoute) UpdatedAt() time.Time   { return r.updatedAt }
func (r *SyncRoute) DeletedAt() *time.Time  { return r.deletedAt }

func (r *SyncRoute) NeedsWebhookSecret() bool {
	return r.trigger == SyncRouteTriggerWebhook || r.trigger == SyncRouteTriggerBoth
}

// AttachWebhookSecretCiphertext stores opaque ciphertext on create (no rotate event).
func (r *SyncRoute) AttachWebhookSecretCiphertext(ciphertext []byte, now time.Time) error {
	if r.deletedAt != nil {
		return ErrSyncRouteRemoved
	}
	if len(ciphertext) == 0 {
		return ErrEmptySecret
	}
	r.webhookSecretCipher = append([]byte(nil), ciphertext...)
	r.updatedAt = now.UTC()
	return nil
}

// RotateWebhookSecretCiphertext replaces ciphertext and records a rotate fact.
func (r *SyncRoute) RotateWebhookSecretCiphertext(ciphertext []byte, now time.Time) error {
	if err := r.AttachWebhookSecretCiphertext(ciphertext, now); err != nil {
		return err
	}
	r.record(EventSyncRouteWebhookSecretRotated, r.id, now.UTC(), map[string]any{
		"workspace_id": r.workspaceID.String(),
	})
	return nil
}

type UpdateSyncRouteParams struct {
	Name               *string
	SourceConnectionID *uuid.UUID
	Trigger            *SyncRouteTrigger
	ResourceTypes      []string
	Schedule           map[string]any
	Filter             map[string]any
	IdempotencyRule    map[string]any
	Activities         []ActivityGroup
	Reverse            *ReverseConfig
	RetryPolicy        map[string]any
	ClearReverse       bool
	Now                time.Time
}

func (r *SyncRoute) Update(p UpdateSyncRouteParams) error {
	if r.deletedAt != nil {
		return ErrSyncRouteRemoved
	}
	if r.status == SyncRouteStatusEnabled {
		return ErrSyncRouteNotEditable
	}

	if p.Name != nil {
		name := strings.TrimSpace(*p.Name)
		if name == "" || len(name) > 255 {
			return ErrInvalidName
		}
		r.name = name
	}
	if p.SourceConnectionID != nil {
		if *p.SourceConnectionID == uuid.Nil {
			return ErrInvalidName
		}
		r.sourceConnectionID = *p.SourceConnectionID
	}
	if p.Trigger != nil {
		if _, err := ParseSyncRouteTrigger(string(*p.Trigger)); err != nil {
			return err
		}
		r.trigger = *p.Trigger
	}
	if p.ResourceTypes != nil {
		types, err := normalizeResourceTypes(p.ResourceTypes)
		if err != nil {
			return err
		}
		r.resourceTypes = types
	}
	if p.Schedule != nil {
		r.schedule = p.Schedule
	}
	if p.Filter != nil {
		r.filter = p.Filter
	}
	if p.IdempotencyRule != nil {
		r.idempotencyRule = p.IdempotencyRule
	}
	if p.Activities != nil {
		activities, err := normalizeActivities(p.Activities)
		if err != nil {
			return err
		}
		r.activities = activities
	}
	if p.ClearReverse {
		r.reverse = nil
	} else if p.Reverse != nil {
		reverse, err := normalizeReverse(p.Reverse)
		if err != nil {
			return err
		}
		r.reverse = reverse
	}
	if p.RetryPolicy != nil {
		r.retryPolicy = p.RetryPolicy
	}

	r.updatedAt = p.Now.UTC()
	r.record(EventSyncRouteUpdated, r.id, p.Now.UTC(), map[string]any{
		"workspace_id": r.workspaceID.String(),
		"name":         r.name,
	})
	return nil
}

func (r *SyncRoute) Enable(now time.Time) error {
	if r.deletedAt != nil {
		return ErrSyncRouteRemoved
	}
	if r.status == SyncRouteStatusEnabled {
		return nil
	}
	if r.status != SyncRouteStatusDraft && r.status != SyncRouteStatusDisabled {
		return ErrInvalidTransition
	}
	if err := r.validateReady(); err != nil {
		return err
	}
	r.status = SyncRouteStatusEnabled
	r.updatedAt = now.UTC()
	r.record(EventSyncRouteEnabled, r.id, now.UTC(), map[string]any{
		"workspace_id": r.workspaceID.String(),
	})
	return nil
}

func (r *SyncRoute) Disable(now time.Time) error {
	if r.deletedAt != nil {
		return ErrSyncRouteRemoved
	}
	if r.status != SyncRouteStatusEnabled {
		return ErrInvalidTransition
	}
	r.status = SyncRouteStatusDisabled
	r.updatedAt = now.UTC()
	r.record(EventSyncRouteDisabled, r.id, now.UTC(), map[string]any{
		"workspace_id": r.workspaceID.String(),
	})
	return nil
}

func (r *SyncRoute) SoftDelete(now time.Time) error {
	if r.deletedAt != nil {
		return ErrSyncRouteRemoved
	}
	if r.status == SyncRouteStatusEnabled {
		return ErrInvalidTransition
	}
	ts := now.UTC()
	r.deletedAt = &ts
	r.updatedAt = ts
	r.record(EventSyncRouteDeleted, r.id, ts, map[string]any{
		"workspace_id": r.workspaceID.String(),
	})
	return nil
}

func (r *SyncRoute) validateReady() error {
	if len(r.resourceTypes) == 0 {
		return ErrInvalidResourceTypes
	}
	if len(r.activities) == 0 {
		return ErrInvalidActivityGroup
	}
	if r.NeedsWebhookSecret() && len(r.webhookSecretCipher) == 0 {
		return ErrWebhookSecretRequired
	}
	if r.trigger == SyncRouteTriggerSchedule || r.trigger == SyncRouteTriggerBoth {
		if len(r.schedule) == 0 {
			return ErrInvalidSchedule
		}
	}
	return nil
}

func normalizeResourceTypes(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, ErrInvalidResourceTypes
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		v := strings.TrimSpace(strings.ToLower(raw))
		if v == "" || len(v) > 100 {
			return nil, ErrInvalidResourceTypes
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, ErrInvalidResourceTypes
	}
	return out, nil
}

func normalizeActivities(in []ActivityGroup) ([]ActivityGroup, error) {
	if len(in) == 0 {
		return nil, ErrInvalidActivityGroup
	}
	out := make([]ActivityGroup, 0, len(in))
	for _, g := range in {
		mode, err := ParseActivityGroupMode(string(g.Mode))
		if err != nil {
			return nil, err
		}
		if len(g.Steps) == 0 {
			return nil, ErrInvalidActivityGroup
		}
		steps := make([]ActivityStep, 0, len(g.Steps))
		for _, s := range g.Steps {
			capCode := strings.TrimSpace(s.Capability)
			if s.DestinationConnectionID == uuid.Nil || capCode == "" || len(capCode) > 150 {
				return nil, ErrInvalidActivityStep
			}
			steps = append(steps, ActivityStep{
				DestinationConnectionID: s.DestinationConnectionID,
				Capability:              capCode,
				MappingKey:              strings.TrimSpace(s.MappingKey),
			})
		}
		out = append(out, ActivityGroup{Mode: mode, Steps: steps})
	}
	return out, nil
}

func normalizeReverse(in *ReverseConfig) (*ReverseConfig, error) {
	if in == nil {
		return nil, nil
	}
	capCode := strings.TrimSpace(in.Capability)
	if capCode == "" || len(capCode) > 150 {
		return nil, ErrInvalidReverseConfig
	}
	on, err := ParseReverseOn(in.On)
	if err != nil {
		return nil, err
	}
	return &ReverseConfig{Capability: capCode, On: on}, nil
}

// SyncRouteWatermark is the poll cursor for one SyncRoute + resource_type.
type SyncRouteWatermark struct {
	SyncRouteID  uuid.UUID
	ResourceType string
	Cursor       map[string]any
	UpdatedAt    time.Time
}

func NewSyncRouteWatermark(syncRouteID uuid.UUID, resourceType string, cursor map[string]any, now time.Time) (*SyncRouteWatermark, error) {
	resourceType = strings.TrimSpace(strings.ToLower(resourceType))
	if syncRouteID == uuid.Nil || resourceType == "" || len(resourceType) > 100 {
		return nil, ErrInvalidResourceTypes
	}
	if cursor == nil {
		cursor = map[string]any{}
	}
	return &SyncRouteWatermark{
		SyncRouteID:  syncRouteID,
		ResourceType: resourceType,
		Cursor:       cursor,
		UpdatedAt:    now.UTC(),
	}, nil
}
