package domain

// DefaultRequestPipelineSpec is the built-in Canonical → Canonical normalization applied to
// invoice *create/publish* Intent payloads before they reach a Connector Runtime. Orchestration
// selects this spec only for create-like capabilities (see TransformerAdapter); get/update_status
// Intents use an identity transform so ValidateRequired does not reject non-document payloads.
// Every Operation is a no-op on fields that are absent — only ValidateRequired can fail a Document.
func DefaultRequestPipelineSpec() []OperationSpec {
	return []OperationSpec{
		{Type: OpTypeRenameField, Params: map[string]any{"from": "buyer_name", "to": "customer.name"}},
		{Type: OpTypeConvertType, Params: map[string]any{"path": "total", "to": string(ConvertToNumber)}},
		{Type: OpTypeNormalizeCurrency, Params: map[string]any{"path": "currency"}},
		{Type: OpTypeNormalizeTimezone, Params: map[string]any{"path": "issue_date", "target_tz": "UTC"}},
		{Type: OpTypeSetDefault, Params: map[string]any{"path": "status", "value": "pending"}},
		{Type: OpTypeValidateRequired, Params: map[string]any{"paths": []string{"invoice_number", "issue_date"}}},
	}
}

// DefaultResponsePipelineSpec normalizes a Connector Runtime's Canonical response. It is
// intentionally lighter than the request pipeline (no rename/defaults/validation): a Connector
// already returns its own Canonical Resource, so Transformation only re-asserts the
// platform-wide normalization rules (currency, timezone) documented in docs/06 §4.
func DefaultResponsePipelineSpec() []OperationSpec {
	return []OperationSpec{
		{Type: OpTypeNormalizeCurrency, Params: map[string]any{"path": "currency"}},
		{Type: OpTypeNormalizeTimezone, Params: map[string]any{"path": "issue_date", "target_tz": "UTC"}},
	}
}
