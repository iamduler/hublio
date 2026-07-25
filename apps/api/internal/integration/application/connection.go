package application

import (
	"context"
	"encoding/json"
	"time"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

type CreateConnectionInput struct {
	WorkspaceID         uuid.UUID
	ConnectorID         uuid.UUID
	Name                string
	IsDefault           bool
	Description         string
	Environment         string
	Config              map[string]any
	RetryPolicy         map[string]any
	TimeoutSeconds      int
	CredentialType      domain.CredentialType
	Secret              map[string]any
	CredentialExpiresAt *time.Time
	ActorUserID         uuid.UUID
}

type CreateConnectionResult struct {
	Connection *domain.Connection
	Credential *domain.Credential
}

// CreateConnection creates a Draft Connection with an initial encrypted Credential.
// The Connector must be Enabled; the returned Credential never carries plaintext.
func (s *Services) CreateConnection(ctx context.Context, in CreateConnectionInput) (*CreateConnectionResult, error) {
	connector, err := s.Connectors.FindByID(ctx, in.ConnectorID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !connector.IsUsable() {
		return nil, apperr.New("connector is not enabled", apperr.ErrCodeConflict)
	}

	now := s.clock().Now()

	connID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate connection id", apperr.ErrCodeInternal)
	}
	conn, err := domain.NewConnection(connID, in.WorkspaceID, connector.ID(), in.Name, in.IsDefault, in.Description, in.Environment, in.Config, in.RetryPolicy, in.TimeoutSeconds, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connections.Save(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}

	cred, err := s.createCredential(ctx, conn, in.CredentialType, in.Secret, in.CredentialExpiresAt, in.ActorUserID, now)
	if err != nil {
		return nil, err
	}

	conn.SetActiveCredential(cred.ID(), now)
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}

	return &CreateConnectionResult{Connection: conn, Credential: cred}, nil
}

func (s *Services) createCredential(
	ctx context.Context,
	conn *domain.Connection,
	credType domain.CredentialType,
	secret map[string]any,
	expiresAt *time.Time,
	createdBy uuid.UUID,
	now time.Time,
) (*domain.Credential, error) {
	ciphertext, err := s.encryptSecret(secret)
	if err != nil {
		return nil, err
	}
	credID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate credential id", apperr.ErrCodeInternal)
	}
	cred, err := domain.NewCredential(credID, conn.ID(), credType, 1, ciphertext, expiresAt, createdBy, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Credentials.Save(ctx, cred); err != nil {
		return nil, mapRepoErr(err)
	}
	return cred, nil
}

func (s *Services) encryptSecret(secret map[string]any) ([]byte, error) {
	if s.Secrets == nil {
		return nil, apperr.New("secret encryptor not configured", apperr.ErrCodeInternal)
	}
	plaintext, err := json.Marshal(secret)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to encode credential secret", apperr.ErrCodeBadRequest)
	}
	ciphertext, err := s.Secrets.Encrypt(plaintext)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to encrypt credential secret", apperr.ErrCodeInternal)
	}
	return ciphertext, nil
}

func (s *Services) decryptSecret(ciphertext []byte) (map[string]any, error) {
	if s.Secrets == nil {
		return nil, apperr.New("secret encryptor not configured", apperr.ErrCodeInternal)
	}
	plaintext, err := s.Secrets.Decrypt(ciphertext)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to decrypt credential secret", apperr.ErrCodeInternal)
	}
	secret := map[string]any{}
	if len(plaintext) > 0 {
		if err := json.Unmarshal(plaintext, &secret); err != nil {
			return nil, apperr.Wrap(err, "failed to decode credential secret", apperr.ErrCodeInternal)
		}
	}
	return secret, nil
}

// VerifyConnection runs StartVerify -> Connector Runtime Verify -> MarkVerified|MarkVerificationFailed.
// It never returns an error solely because verification failed; callers should inspect the
// returned Connection's Status().
func (s *Services) VerifyConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
	conn, err := s.findWorkspaceConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return nil, err
	}

	now := s.clock().Now()
	if err := conn.StartVerify(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}

	verifyErr := s.runVerify(ctx, conn)

	now = s.clock().Now()
	if verifyErr != nil {
		if err := conn.MarkVerificationFailed(verifyFailureReason(verifyErr), now); err != nil {
			return nil, mapDomainErr(err)
		}
	} else {
		if err := conn.MarkVerified(now); err != nil {
			return nil, mapDomainErr(err)
		}
	}
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}
	return conn, nil
}

func (s *Services) runVerify(ctx context.Context, conn *domain.Connection) error {
	connector, err := s.Connectors.FindByID(ctx, conn.ConnectorID())
	if err != nil {
		return err
	}
	runtime, err := s.Runtimes.Resolve(connector.Code())
	if err != nil {
		return err
	}
	secretMap := map[string]any{}
	if cred, err := s.Credentials.FindActiveByConnection(ctx, conn.ID()); err == nil {
		secretMap, err = s.decryptSecret(cred.EncryptedSecret())
		if err != nil {
			return err
		}
	}
	if err := runtime.Verify(ctx, domain.VerifyInput{
		ConnectionID: conn.ID(),
		Config:       conn.Config(),
		Secret:       secretMap,
	}); err != nil {
		return mapRuntimeErr(err)
	}
	return nil
}

func (s *Services) DisableConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
	conn, err := s.findWorkspaceConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return nil, err
	}
	if err := conn.Disable(s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}
	return conn, nil
}

func (s *Services) EnableConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
	conn, err := s.findWorkspaceConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return nil, err
	}
	if err := conn.Enable(s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}
	return conn, nil
}

func (s *Services) GetConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
	return s.findWorkspaceConnection(ctx, workspaceID, connectionID)
}

func (s *Services) ListConnections(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Connection, error) {
	list, err := s.Connections.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

type RotateCredentialInput struct {
	WorkspaceID    uuid.UUID
	ConnectionID   uuid.UUID
	CredentialType domain.CredentialType
	Secret         map[string]any
	ExpiresAt      *time.Time
	ActorUserID    uuid.UUID
}

// RotateCredential revokes the current active Credential (if any) and creates a new one,
// pointing the Connection's active_credential_id at it. Plaintext is never returned.
func (s *Services) RotateCredential(ctx context.Context, in RotateCredentialInput) (*domain.Credential, error) {
	conn, err := s.findWorkspaceConnection(ctx, in.WorkspaceID, in.ConnectionID)
	if err != nil {
		return nil, err
	}

	var previous *domain.Credential
	if existing, err := s.Credentials.FindActiveByConnection(ctx, conn.ID()); err == nil {
		previous = existing
	} else if err != domain.ErrNotFound {
		return nil, mapRepoErr(err)
	}

	ciphertext, err := s.encryptSecret(in.Secret)
	if err != nil {
		return nil, err
	}

	now := s.clock().Now()
	newID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate credential id", apperr.ErrCodeInternal)
	}

	next, err := domain.RotateCredential(newID, previous, conn.ID(), in.CredentialType, ciphertext, in.ExpiresAt, in.ActorUserID, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	if previous != nil {
		if err := s.Credentials.Update(ctx, previous); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	if err := s.Credentials.Save(ctx, next); err != nil {
		return nil, mapRepoErr(err)
	}

	conn.SetActiveCredential(next.ID(), now)
	if err := s.Connections.Update(ctx, conn); err != nil {
		return nil, mapRepoErr(err)
	}

	return next, nil
}

func (s *Services) findWorkspaceConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (*domain.Connection, error) {
	conn, err := s.Connections.FindByID(ctx, connectionID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if conn.WorkspaceID() != workspaceID {
		return nil, apperr.New("connection not found", apperr.ErrCodeNotFound)
	}
	return conn, nil
}
