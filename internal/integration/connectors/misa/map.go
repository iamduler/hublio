package misa

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Credentials extracted from decrypted Connection secret + config. Never logged.
type credentials struct {
	AppID    string
	Username string
	Password string
	TaxCode  string
}

type connectionSettings struct {
	BaseURL       string
	InvSeries     string
	SignType      int
	PaymentMethod string
	ExchangeRate  float64
}

func parseCredentials(config, secret map[string]any) (credentials, error) {
	cred := credentials{
		AppID:    stringField(secret, "app_id", "appid", "appId"),
		Username: stringField(secret, "username", "user"),
		Password: stringField(secret, "password"),
		TaxCode:  stringField(secret, "tax_code", "taxcode", "taxCode"),
	}
	if cred.TaxCode == "" {
		cred.TaxCode = stringField(config, "tax_code", "taxcode", "taxCode")
	}
	if cred.AppID == "" || cred.Username == "" || cred.Password == "" || cred.TaxCode == "" {
		return credentials{}, ErrMissingCredentials
	}
	return cred, nil
}

func parseSettings(config map[string]any) connectionSettings {
	signType := intField(config, "sign_type", "signType")
	if signType == 0 {
		// SignType 2 = HSM / server-side publish without local SignService tool.
		signType = 2
	}
	rate := floatField(config, "exchange_rate", "exchangeRate")
	if rate == 0 {
		rate = 1
	}
	payment := stringField(config, "payment_method", "paymentMethodName", "payment_method_name")
	if payment == "" {
		payment = "TM/CK"
	}
	return connectionSettings{
		BaseURL:       stringField(config, "base_url", "baseUrl"),
		InvSeries:     stringField(config, "inv_series", "invSeries", "invoice_series"),
		SignType:      signType,
		PaymentMethod: payment,
		ExchangeRate:  rate,
	}
}

// toProviderInvoice maps a Canonical Invoice document into a meInvoice InvoiceData DTO.
func toProviderInvoice(payload map[string]any, settings connectionSettings) (invoiceData, error) {
	if payload == nil {
		return invoiceData{}, ErrInvalidPayload
	}
	invoiceNumber := stringField(payload, "invoice_number", "invoiceNumber")
	issueDate := stringField(payload, "issue_date", "issueDate")
	if invoiceNumber == "" || issueDate == "" {
		return invoiceData{}, fmt.Errorf("%w: invoice_number and issue_date are required", ErrInvalidPayload)
	}

	invDate, err := formatInvDate(issueDate)
	if err != nil {
		return invoiceData{}, fmt.Errorf("%w: issue_date: %v", ErrInvalidPayload, err)
	}

	currency := strings.ToUpper(stringField(payload, "currency"))
	if currency == "" {
		currency = "VND"
	}

	customer, _ := payload["customer"].(map[string]any)
	buyerName := stringField(customer, "name", "legal_name", "legalName")
	if buyerName == "" {
		buyerName = stringField(payload, "buyer_name", "customer_name")
	}
	if buyerName == "" {
		return invoiceData{}, fmt.Errorf("%w: customer.name is required", ErrInvalidPayload)
	}

	items, _ := payload["items"].([]any)
	if len(items) == 0 {
		return invoiceData{}, fmt.Errorf("%w: items are required", ErrInvalidPayload)
	}

	details := make([]originalInvoiceDetail, 0, len(items))
	taxBuckets := map[string]*taxRateInfo{}
	var subtotal, vatTotal float64

	for i, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			return invoiceData{}, fmt.Errorf("%w: items[%d] must be an object", ErrInvalidPayload, i)
		}
		qty := floatField(item, "quantity", "qty")
		if qty == 0 {
			qty = 1
		}
		unitPrice := floatField(item, "unit_price", "unitPrice", "price")
		amount := floatField(item, "amount", "line_total")
		if amount == 0 {
			amount = qty * unitPrice
		}
		vatRate := stringField(item, "vat_rate", "vatRate", "tax_rate")
		if vatRate == "" {
			vatRate = "10%"
		}
		vatAmount := floatField(item, "vat_amount", "vatAmount")
		if vatAmount == 0 {
			vatAmount = amount * vatRateFraction(vatRate)
		}
		line := originalInvoiceDetail{
			ItemType:           1,
			LineNumber:         i + 1,
			SortOrder:          i + 1,
			ItemCode:           stringField(item, "code", "item_code", "sku"),
			ItemName:           stringField(item, "name", "item_name"),
			UnitName:           stringField(item, "unit", "unit_name"),
			Quantity:           qty,
			UnitPrice:          unitPrice,
			DiscountRate:       floatField(item, "discount_rate", "discountRate"),
			AmountOC:           amount,
			Amount:             amount,
			AmountWithoutVATOC: amount,
			AmountWithoutVAT:   amount,
			VATRateName:        vatRate,
			VATAmountOC:        vatAmount,
			VATAmount:          vatAmount,
		}
		if line.ItemName == "" {
			return invoiceData{}, fmt.Errorf("%w: items[%d].name is required", ErrInvalidPayload, i)
		}
		details = append(details, line)
		subtotal += amount
		vatTotal += vatAmount
		if b, ok := taxBuckets[vatRate]; ok {
			b.AmountWithoutVATOC += amount
			b.VATAmountOC += vatAmount
		} else {
			taxBuckets[vatRate] = &taxRateInfo{
				VATRateName:        vatRate,
				AmountWithoutVATOC: amount,
				VATAmountOC:        vatAmount,
			}
		}
	}

	if v := floatField(payload, "subtotal"); v > 0 {
		subtotal = v
	}
	total := floatField(payload, "total")
	if total == 0 {
		total = subtotal + vatTotal
	}
	if v := floatField(payload, "tax_total", "vat_total"); v > 0 {
		vatTotal = v
	}

	taxes := make([]taxRateInfo, 0, len(taxBuckets))
	for _, b := range taxBuckets {
		taxes = append(taxes, *b)
	}

	refID := stringField(payload, "id", "ref_id", "refId")
	if refID == "" {
		refID = uuid.NewString()
	}

	series := settings.InvSeries
	if series == "" {
		series = stringField(payload, "inv_series", "series")
	}
	if series == "" {
		return invoiceData{}, fmt.Errorf("%w: inv_series config is required", ErrMissingConfig)
	}

	return invoiceData{
		RefID:                   refID,
		InvSeries:               series,
		InvDate:                 invDate,
		CurrencyCode:            currency,
		ExchangeRate:            settings.ExchangeRate,
		PaymentMethodName:       settings.PaymentMethod,
		BuyerLegalName:          buyerName,
		BuyerTaxCode:            stringField(customer, "tax_code", "taxCode"),
		BuyerAddress:            stringField(customer, "address"),
		BuyerCode:               stringField(customer, "code", "customer_code"),
		BuyerPhoneNumber:        stringField(customer, "phone", "mobile", "phone_number"),
		BuyerEmail:              stringField(customer, "email"),
		BuyerFullName:           stringField(customer, "full_name", "fullName"),
		TotalSaleAmountOC:       subtotal,
		TotalSaleAmount:         subtotal,
		TotalAmountWithoutVATOC: subtotal,
		TotalAmountWithoutVAT:   subtotal,
		TotalVATAmountOC:        vatTotal,
		TotalVATAmount:          vatTotal,
		TotalDiscountAmountOC:   0,
		TotalDiscountAmount:     0,
		TotalAmountOC:           total,
		TotalAmount:             total,
		OriginalInvoiceDetail:   details,
		TaxRateInfo:             taxes,
	}, nil
}

func toCanonicalResponse(payload map[string]any, result *publishResponse) map[string]any {
	out := cloneMap(payload)
	out["status"] = "published"
	if result == nil {
		return out
	}
	if len(result.PublishInvoiceResult) > 0 {
		r := result.PublishInvoiceResult[0]
		if r.InvNo != "" {
			out["invoice_number"] = r.InvNo
		}
		if r.InvDate != "" {
			out["issue_date"] = r.InvDate
		}
		if r.TransactionID != "" {
			out["provider_transaction_id"] = r.TransactionID
		}
		if r.RefID != "" {
			out["id"] = r.RefID
		}
	}
	return out
}

func responseMetadata(result *publishResponse) map[string]any {
	meta := map[string]any{"connector": Code}
	if result == nil {
		return meta
	}
	if len(result.PublishInvoiceResult) > 0 {
		r := result.PublishInvoiceResult[0]
		meta["transaction_id"] = r.TransactionID
		meta["inv_series"] = r.InvSeries
		meta["inv_no"] = r.InvNo
	}
	return meta
}

func formatInvDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 10 && raw[4] == '-' && raw[7] == '-' {
		// Already yyyy-MM-dd or RFC3339 prefix.
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return t.Format("2006-01-02"), nil
		}
		return raw[:10], nil
	}
	layouts := []string{"2006-01-02", time.RFC3339, "02/01/2006"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("unrecognized date %q", raw)
}

func vatRateFraction(rate string) float64 {
	rate = strings.TrimSpace(strings.TrimSuffix(rate, "%"))
	switch rate {
	case "0", "KCT", "KKKNT":
		return 0
	case "5":
		return 0.05
	case "8":
		return 0.08
	case "10":
		return 0.10
	default:
		var f float64
		_, _ = fmt.Sscanf(rate, "%f", &f)
		if f > 1 {
			return f / 100
		}
		return f
	}
}

func stringField(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			case fmt.Stringer:
				if s := strings.TrimSpace(t.String()); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func floatField(m map[string]any, keys ...string) float64 {
	if m == nil {
		return 0
	}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return t
			case float32:
				return float64(t)
			case int:
				return float64(t)
			case int64:
				return float64(t)
			case string:
				var f float64
				_, _ = fmt.Sscanf(strings.TrimSpace(t), "%f", &f)
				return f
			}
		}
	}
	return 0
}

func intField(m map[string]any, keys ...string) int {
	return int(floatField(m, keys...))
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
