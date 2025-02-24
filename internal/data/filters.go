package data

import (
	"greenlight/internal/validator"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, filter Filters) {

	v.Check(filter.Page < 1, "page", "page number less than 1")
	v.Check(filter.Page > 10000000, "page", "invalid page number")
	v.Check(filter.PageSize < 1 || filter.PageSize > 100, "pagesize", "page size can not be more thann 100 or invalid")

	v.Check(validator.PermittedValue(filter.Sort, filter.SortSafeList...), "sort", "invalid sort filter")
}
