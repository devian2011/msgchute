package pagination

import (
	"math"
	"net/http"
	"strconv"
)

type PageData struct {
	CurrentPage uint64 `json:"current_page"`
	PerPage     uint64 `json:"per_page"`
	Total       uint64 `json:"total"`
	TotalPages  uint64 `json:"total_pages"`
}

func GetPageDataFromRequest(r *http.Request) (page, perPage uint64) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, parseErr := strconv.ParseUint(pageStr, 10, 64)
	if parseErr != nil {
		page = 1
	}

	perPageStr := r.URL.Query().Get("perPage")

	perPage, parseErr = strconv.ParseUint(perPageStr, 10, 64)
	if parseErr != nil {
		perPage = 20
	}

	return page, perPage
}

func GetLimitsByPage(page, perPage uint64) (limit, offset uint64) {
	if page == 0 {
		page = 1
	}

	page -= 1

	return perPage, page * perPage
}

func GetPageCount(perPage, itemCount int) int {
	if perPage == 0 {
		return 0
	}
	return int(math.Ceil(float64(itemCount) / float64(perPage)))
}
