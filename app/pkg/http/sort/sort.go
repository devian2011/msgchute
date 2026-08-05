package sort

import (
	"net/http"
)

type Sort struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

func GetSortFromRequest(r *http.Request) Sort {
	sortBy := r.URL.Query().Get("sortBy")
	if len(sortBy) == 0 {
		sortBy = "id"
	}
	sortOrder := r.URL.Query().Get("order")
	if len(sortOrder) == 0 {
		sortOrder = "DESC"
	}

	return Sort{Field: sortBy, Order: sortOrder}
}
