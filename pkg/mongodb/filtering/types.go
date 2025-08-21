package filtering

import (
	"encoding/json"
)

type FilterableField struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`       // string, number, date, boolean
	Operations []string `json:"operations"` // equals, contains, between, gt, lt vs.
} //@name FilterableField

type SortableField struct {
	Name string `json:"name"`
} //@name SortableField

type FilterMetadata struct {
	FilterableFields []FilterableField `json:"filterableFields"`
	SortableFields   []SortableField   `json:"sortableFields"`
} //@name FilterMetadata

type PaginationInfo struct {
	TotalCount  int `json:"totalCount"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPages  int `json:"totalPages"`
} //@name PaginationInfo

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type BasePaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
} //@name BasePaginationRequest

type BaseSortingRequest struct {
	SortBy   string        `form:"sortBy" binding:"omitempty"`
	SortType SortDirection `form:"sortType" binding:"omitempty,oneof=asc desc"`
} //@name BaseSortingRequest

type FilterRequest struct {
	FiltersJSON string `form:"filters"`
	SortBy      string `form:"sortBy"`
	SortOrder   string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (fr *FilterRequest) GetFilters() ([]Filter, error) {
	if fr.FiltersJSON == "" {
		return nil, nil
	}
	var filters []Filter
	err := json.Unmarshal([]byte(fr.FiltersJSON), &filters)
	return filters, err
}

type Filter struct {
	Field     string      `json:"field"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value,omitempty"`
	ValueFrom interface{} `json:"valueFrom,omitempty"`
	ValueTo   interface{} `json:"valueTo,omitempty"`
}
