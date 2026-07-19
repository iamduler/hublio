package domain

// DefaultRequestPipelineSpec is the built-in Canonical → Canonical normalization applied to
// invoice-like Intent payloads before they reach a Connector Runtime. Every Operation is a
// no-op on fields that are absent, so callers that apply this spec unconditionally never break
// payloads that simply do not carry these Canonical invoice fields — only the required-field
// validation can fail a Document, and it is limited to the two fields every Canonical invoice
// must carry.
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
