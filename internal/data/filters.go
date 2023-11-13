package data

import (
	"greenlight.twd.net/internal/validator"
	"math"
	"strings"
)

// Metadata struct for holding the pagination metadata
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"count,omitempty"`
}

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// calculateMetadata function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values.
// Note that the last page value is calculated using the math.Ceil() function,
// which rounds up a float to th nearest integer. So, for example, if there were
// 12 records in total and a page size of 5, the last page value would be math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// limit puts checks around the page size so we can control the number of items returned
func (f Filters) limit() int {
	return f.PageSize
}

// offset determines how many records we want to skip before returning our dataset
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// sortColumn checks the client-provided sort field matches one of the entries in the Filters.SortSafelist list
// and if it does, extract the column name from the Sort field by stripping the leading hyphen character (if one exists)
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection returns direction (ASC, DESC) depending on the prefix character of the sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
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
