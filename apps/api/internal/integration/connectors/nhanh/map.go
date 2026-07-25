package nhanh

import (
	"fmt"
	"strconv"
)

func billToCanonical(b *retailBill) map[string]any {
	out := map[string]any{
		"id":             strconv.FormatInt(b.ID, 10),
		"invoice_number": strconv.FormatInt(b.ID, 10),
		"issue_date":     b.Date,
		"currency":       "VND",
		"status":         "issued",
		"total":          b.Money,
	}
	if b.OrderID != 0 {
		out["order_id"] = strconv.FormatInt(b.OrderID, 10)
	}
	if b.Customer != nil {
		out["customer"] = map[string]any{
			"id":      strconv.FormatInt(b.Customer.ID, 10),
			"name":    b.Customer.Name,
			"phone":   b.Customer.Mobile,
			"address": b.Customer.Address,
		}
	}
	items := make([]any, 0, len(b.Products))
	for _, p := range b.Products {
		items = append(items, map[string]any{
			"code":       p.Code,
			"name":       p.Name,
			"quantity":   p.Quantity,
			"unit_price": p.Price,
			"vat_rate":   fmt.Sprintf("%.0f%%", p.VAT),
		})
	}
	if len(items) > 0 {
		out["items"] = items
	}
	return out
}
