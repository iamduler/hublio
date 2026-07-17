ALTER TABLE connections DROP CONSTRAINT IF EXISTS connections_active_credential_id_fkey;

DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS connector_capabilities;
DROP TABLE IF EXISTS connectors;

DROP TYPE IF EXISTS credential_status;
DROP TYPE IF EXISTS credential_type;
DROP TYPE IF EXISTS connection_status;
DROP TYPE IF EXISTS connector_capability_status;
DROP TYPE IF EXISTS connector_category;
DROP TYPE IF EXISTS connector_status;
