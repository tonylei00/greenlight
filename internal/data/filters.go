package data

import (
	"strings"

	"greenlight.tlei.net/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f *Filters) {
	v.Check(f.Page >= 1 && f.Page <= 10_000_000, "page", "value must be between 1 and 10,000,000")
	v.Check(f.PageSize >= 1 && f.PageSize <= 100, "page_size", "value must be between 1 and 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, safe := range f.SortSafeList {
		if f.Sort == safe {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter:" + f.Sort)
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
