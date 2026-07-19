package misa

// Provider DTOs live only in this package. They must never be returned from Runtime methods
// as typed values — only mapped into Canonical map[string]any payloads / Metadata.

type tokenRequest struct {
	AppID    string `json:"appid"`
	TaxCode  string `json:"taxcode"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type apiEnvelope struct {
	Success   bool   `json:"Success"`
	Data      any    `json:"Data"`
	ErrorCode string `json:"ErrorCode"`
	Errors    any    `json:"Errors"`
}

// Lowercase variants appear in some meInvoice Open API responses.
type publishResponse struct {
	Success              bool                   `json:"success"`
	ErrorCode            *string                `json:"errorCode"`
	DescriptionErrorCode *string                `json:"descriptionErrorCode"`
	CreateInvoiceResult  []map[string]any       `json:"createInvoiceResult"`
	PublishInvoiceResult []publishInvoiceResult `json:"publishInvoiceResult"`
}

type publishInvoiceResult struct {
	TransactionID string `json:"TransactionID"`
	ErrorCode     string `json:"ErrorCode"`
	InvNo         string `json:"InvNo"`
	InvSeries     string `json:"InvSeries"`
	InvDate       string `json:"InvDate"`
	RefID         string `json:"RefID"`
}

type createInvoiceRequest struct {
	SignType           int           `json:"SignType"`
	InvoiceData        []invoiceData `json:"InvoiceData"`
	PublishInvoiceData any           `json:"PublishInvoiceData"`
}

type invoiceData struct {
	RefID                   string                  `json:"RefID"`
	InvSeries               string                  `json:"InvSeries"`
	InvDate                 string                  `json:"InvDate"`
	CurrencyCode            string                  `json:"CurrencyCode"`
	ExchangeRate            float64                 `json:"ExchangeRate"`
	PaymentMethodName       string                  `json:"PaymentMethodName"`
	BuyerLegalName          string                  `json:"BuyerLegalName"`
	BuyerTaxCode            string                  `json:"BuyerTaxCode,omitempty"`
	BuyerAddress            string                  `json:"BuyerAddress,omitempty"`
	BuyerCode               string                  `json:"BuyerCode,omitempty"`
	BuyerPhoneNumber        string                  `json:"BuyerPhoneNumber,omitempty"`
	BuyerEmail              string                  `json:"BuyerEmail,omitempty"`
	BuyerFullName           string                  `json:"BuyerFullName,omitempty"`
	TotalSaleAmountOC       float64                 `json:"TotalSaleAmountOC"`
	TotalSaleAmount         float64                 `json:"TotalSaleAmount"`
	TotalAmountWithoutVATOC float64                 `json:"TotalAmountWithoutVATOC"`
	TotalAmountWithoutVAT   float64                 `json:"TotalAmountWithoutVAT"`
	TotalVATAmountOC        float64                 `json:"TotalVATAmountOC"`
	TotalVATAmount          float64                 `json:"TotalVATAmount"`
	TotalDiscountAmountOC   float64                 `json:"TotalDiscountAmountOC"`
	TotalDiscountAmount     float64                 `json:"TotalDiscountAmount"`
	TotalAmountOC           float64                 `json:"TotalAmountOC"`
	TotalAmount             float64                 `json:"TotalAmount"`
	TotalAmountInWords      string                  `json:"TotalAmountInWords,omitempty"`
	OriginalInvoiceDetail   []originalInvoiceDetail `json:"OriginalInvoiceDetail"`
	TaxRateInfo             []taxRateInfo           `json:"TaxRateInfo"`
}

type originalInvoiceDetail struct {
	ItemType           int     `json:"ItemType"`
	LineNumber         int     `json:"LineNumber"`
	SortOrder          int     `json:"SortOrder"`
	ItemCode           string  `json:"ItemCode,omitempty"`
	ItemName           string  `json:"ItemName"`
	UnitName           string  `json:"UnitName,omitempty"`
	Quantity           float64 `json:"Quantity"`
	UnitPrice          float64 `json:"UnitPrice"`
	DiscountRate       float64 `json:"DiscountRate"`
	DiscountAmountOC   float64 `json:"DiscountAmountOC"`
	DiscountAmount     float64 `json:"DiscountAmount"`
	AmountOC           float64 `json:"AmountOC"`
	Amount             float64 `json:"Amount"`
	AmountWithoutVATOC float64 `json:"AmountWithoutVATOC"`
	AmountWithoutVAT   float64 `json:"AmountWithoutVAT"`
	VATRateName        string  `json:"VATRateName"`
	VATAmountOC        float64 `json:"VATAmountOC"`
	VATAmount          float64 `json:"VATAmount"`
}

type taxRateInfo struct {
	VATRateName        string  `json:"VATRateName"`
	AmountWithoutVATOC float64 `json:"AmountWithoutVATOC"`
	VATAmountOC        float64 `json:"VATAmountOC"`
}
