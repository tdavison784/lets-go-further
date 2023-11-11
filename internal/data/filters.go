package data

import "greenlight.twd.net/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "maximum value of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "maximum value of 100")

	// Check the values of sort parameter matches a value in the safelist
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
