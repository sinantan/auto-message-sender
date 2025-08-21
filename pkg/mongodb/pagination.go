package mongodb

import (
	"math"

	"everby/pkg/mongodb/filtering"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CalculatePaginationInfo(totalCount int64, page, pageSize int) *filtering.PaginationInfo {
	if pageSize == 0 {
		pageSize = 50
	}

	if page == 0 {
		page = 1
	}

	return &filtering.PaginationInfo{
		TotalCount:  int(totalCount),
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}
}

func ApplyPagination(opts *options.FindOptions, page, pageSize int) {
	if pageSize == 0 {
		pageSize = 50
	}

	if page == 0 {
		page = 1
	}

	skip := int64((page - 1) * pageSize)
	opts.SetSkip(skip)
	opts.SetLimit(int64(pageSize))
}

func ApplySorting(opts *options.FindOptions, sortBy string, sortType filtering.SortDirection) {
	if sortBy == "" {
		return
	}

	order := 1
	if sortType == filtering.SortDesc {
		order = -1
	}

	opts.SetSort(bson.D{{sortBy, order}})
}
