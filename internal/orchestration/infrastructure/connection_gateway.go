package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	identitydomain "hublio/internal/identity/domain"
	integrationdomain "hublio/internal/integration/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// SecretDecryptor decrypts an already-encrypted Credential secret. The same
// integrationinfra.AESSecretEncryptor instance used by the Integration BC satisfies this
// interface structurally, so no plaintext key material is duplicated.
type SecretDecryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// ConnectionGateway adapts the Integration and Identity Domains into the Orchestration
// Application's ConnectionGateway port. It lives in Infrastructure so that Orchestration's
// Domain/Application never import Integration or Identity types directly.
type ConnectionGateway struct {
	connections integrationdomain.ConnectionRepository
	connectors  integrationdomain.ConnectorRepository
	credentials integrationdomain.CredentialRepository
	workspaces  identitydomain.WorkspaceRepository
	orgs        identitydomain.OrganizationRepository
	secrets     SecretDecryptor
}

func NewConnectionGateway(
	connections integrationdomain.ConnectionRepository,
	connectors integrationdomain.ConnectorRepository,
	credentials integrationdomain.CredentialRepository,
	workspaces identitydomain.WorkspaceRepository,
	orgs identitydomain.OrganizationRepository,
	secrets SecretDecryptor,
) *ConnectionGateway {
	return &ConnectionGateway{
		connections: connections,
		connectors:  connectors,
		credentials: credentials,
		workspaces:  workspaces,
		orgs:        orgs,
		secrets:     secrets,
	}
}

func (g *ConnectionGateway) ResolveForIntent(ctx context.Context, workspaceID, connectionID uuid.UUID) (orchestrationapp.ResolvedConnection, error) {
	ws, err := g.workspaces.FindByID(ctx, workspaceID)
	if err != nil {
		return orchestrationapp.ResolvedConnection{}, apperr.New("workspace not found", apperr.ErrCodeNotFound)
	}
	if !ws.CanExecuteIntents() {
		return orchestrationapp.ResolvedConnection{}, apperr.New("workspace is disabled", apperr.ErrCodeConflict)
	}

	org, err := g.orgs.FindByID(ctx, ws.OrganizationID())
	if err != nil || !org.CanSubmitIntents() {
		return orchestrationapp.ResolvedConnection{}, apperr.New("organization cannot submit intents", apperr.ErrCodeConflict)
	}

	conn, err := g.connections.FindByID(ctx, connectionID)
	if err != nil {
		return orchestrationapp.ResolvedConnection{}, apperr.New("connection not found", apperr.ErrCodeNotFound)
	}
	if conn.WorkspaceID() != workspaceID {
		return orchestrationapp.ResolvedConnection{}, apperr.New("connection not found", apperr.ErrCodeNotFound)
	}
	if !conn.CanExecuteIntents() {
		return orchestrationapp.ResolvedConnection{}, apperr.New("connection is not active", apperr.ErrCodeConflict)
	}

	connector, err := g.connectors.FindByID(ctx, conn.ConnectorID())
	if err != nil {
		return orchestrationapp.ResolvedConnection{}, apperr.New("connector not found", apperr.ErrCodeNotFound)
	}
	if !connector.IsUsable() {
		return orchestrationapp.ResolvedConnection{}, apperr.New("connector is not enabled", apperr.ErrCodeConflict)
	}

	secretMap, err := g.decryptActiveSecret(ctx, conn.ID())
	if err != nil {
		return orchestrationapp.ResolvedConnection{}, err
	}

	return orchestrationapp.ResolvedConnection{
		ConnectionID:   conn.ID(),
		WorkspaceID:    conn.WorkspaceID(),
		ConnectorID:    connector.ID(),
		ConnectorCode:  connector.Code(),
		Config:         conn.Config(),
		Secret:         secretMap,
		TimeoutSeconds: conn.TimeoutSeconds(),
	}, nil
}

func (g *ConnectionGateway) decryptActiveSecret(ctx context.Context, connectionID uuid.UUID) (map[string]any, error) {
	secretMap := map[string]any{}
	cred, err := g.credentials.FindActiveByConnection(ctx, connectionID)
	if err != nil {
		// No active credential yet (e.g. Fake connector without secrets): proceed with none.
		return secretMap, nil
	}
	if err := cred.IsUsable(time.Now().UTC()); err != nil {
		return secretMap, nil
	}
	if g.secrets == nil {
		return secretMap, nil
	}
	plaintext, err := g.secrets.Decrypt(cred.EncryptedSecret())
	if err != nil {
		return nil, apperr.Wrap(err, "failed to decrypt credential secret", apperr.ErrCodeInternal)
	}
	if len(plaintext) == 0 {
		return secretMap, nil
	}
	if err := json.Unmarshal(plaintext, &secretMap); err != nil {
		return nil, apperr.Wrap(err, "failed to decode credential secret", apperr.ErrCodeInternal)
	}
	return secretMap, nil
}
