package nhanh

// Provider DTOs — confined to this package.

type apiResponse struct {
	Code       int    `json:"code"`
	ErrorCode  string `json:"errorCode"`
	Messages   any    `json:"messages"`
	Data       any    `json:"data"`
	BusinessID any    `json:"businessId"`
}

type retailBill struct {
	ID       int64         `json:"id"`
	OrderID  int64         `json:"orderId"`
	Date     string        `json:"date"`
	Customer *billCustomer `json:"customer"`
	Products []billProduct `json:"products"`
	Money    float64       `json:"money"`
}

type billCustomer struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Mobile  string `json:"mobile"`
	Address string `json:"address"`
}

type billProduct struct {
	ID       int64   `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	VAT      float64 `json:"vat"`
}
