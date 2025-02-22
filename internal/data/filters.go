package data

import "greenlight/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, filter Filters) {

	v.Check(filter.Page < 1, "page", "page number less than 1")
	v.Check(filter.Page > 10000000, "page", "invalid page number")
	v.Check(filter.PageSize < 1 || filter.PageSize > 100, "pagesize", "page size can not be more thann 100 or invalid")

	v.Check(validator.PermittedValue(filter.Sort, filter.SortSafeList...), "sort", "invalid sort filter")
}
