package httpx

import "hublio/internal/platform/env"

type Pagination struct {
	Page         int32 `json:"page"`
	Limit        int32 `json:"limit"`
	TotalRecords int32 `json:"total"`
	TotalPages   int32 `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

func NewPagination(page, limit, totalRecords int32) *Pagination {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		envLimit := env.GetIntEnv("LIMIT_ITEMS_PER_PAGE", 10)
		limit = int32(envLimit)
	}

	totalPages := (totalRecords + limit - 1) / limit
	hasNext := page < int32(totalPages)
	hasPrevious := page > 1

	return &Pagination{
		Page:         page,
		Limit:        limit,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		HasNext:      hasNext,
		HasPrevious:  hasPrevious,
	}
}

func NewPaginationResponse(data any, page, limit, totalRecords int32) map[string]any {
	return map[string]any{
		"data":       data,
		"pagination": NewPagination(page, limit, totalRecords),
	}
}
