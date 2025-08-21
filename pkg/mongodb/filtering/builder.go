package filtering

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoFilterBuilder struct {
	Request  *FilterRequest
	Metadata *FilterMetadata
}

func (b *MongoFilterBuilder) BuildFilter() (bson.M, error) {
	filter := bson.M{}

	filters, err := b.Request.GetFilters()
	if err != nil {
		return nil, err
	}

	if filters == nil {
		return filter, nil
	}

	for _, f := range filters {
		if !b.isValidFilter(f) {
			return nil, fmt.Errorf("invalid filter: field '%s' with operation '%s' is not allowed", f.Field, f.Operation)
		}

		if err := b.applyFilter(filter, f); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func (b *MongoFilterBuilder) applyFilter(filter bson.M, f Filter) error {
	switch f.Operation {
	case "between":
		if fromStr, ok := f.ValueFrom.(string); ok && isDateFormat(fromStr) {
			fromDate, err := parseDate(f.ValueFrom)
			if err != nil {
				return err
			}

			toDate, err := parseDate(f.ValueTo)
			if err != nil {
				return err
			}

			filter[f.Field] = bson.M{
				"$gte": fromDate,
				"$lte": toDate,
			}
			return nil
		}

		filter[f.Field] = bson.M{
			"$gte": f.ValueFrom,
			"$lte": f.ValueTo,
		}
	case "contains":
		filter[f.Field] = bson.M{
			"$regex":   f.Value,
			"$options": "i",
		}
	case "equals":
		filter[f.Field] = f.Value
	case "gt":
		filter[f.Field] = bson.M{"$gt": f.Value}
	case "lt":
		filter[f.Field] = bson.M{"$lt": f.Value}
	}
	return nil
}

func isDateFormat(value string) bool {
	if _, err := time.Parse("02-01-2006", value); err == nil {
		return true
	}
	return false
}

func parseDate(value interface{}) (time.Time, error) {
	if str, ok := value.(string); ok {
		t, err := time.Parse("02-01-2006", str)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format. Use DD-MM-YYYY")
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("date must be string")
}

func (b *MongoFilterBuilder) BuildFindOptions() *options.FindOptions {
	opts := options.Find()

	if b.Request.Page > 0 && b.Request.PageSize > 0 {
		skip := int64((b.Request.Page - 1) * b.Request.PageSize)
		opts.SetSkip(skip)
		opts.SetLimit(int64(b.Request.PageSize))
	}

	if b.Request.SortBy != "" {
		order := 1
		if b.Request.SortOrder == "desc" {
			order = -1
		}
		opts.SetSort(bson.D{{b.Request.SortBy, order}})
	}

	return opts
}

func (b *MongoFilterBuilder) isValidFilter(filter Filter) bool {
	for _, field := range b.Metadata.FilterableFields {
		if field.Name == filter.Field {
			for _, op := range field.Operations {
				if op == filter.Operation {
					return true
				}
			}
		}
	}
	return false
}

func ApplyDateRangeFilter(filter bson.M, fieldName, beginDateStr, endDateStr string, dateFormat ...string) error {
	if beginDateStr == "" && endDateStr == "" {
		return nil
	}

	format := "02-01-2006"
	if len(dateFormat) > 0 && dateFormat[0] != "" {
		format = dateFormat[0]
	}

	dateFilter := bson.M{}

	if beginDateStr != "" {
		beginDate, err := time.Parse(format, beginDateStr)
		if err != nil {
			return fmt.Errorf("invalid begin date format: %v", err)
		}
		dateFilter["$gte"] = beginDate
	}

	if endDateStr != "" {
		endDate, err := time.Parse(format, endDateStr)
		if err != nil {
			return fmt.Errorf("invalid end date format: %v", err)
		}
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		dateFilter["$lte"] = endDate
	}

	if len(dateFilter) > 0 {
		filter[fieldName] = dateFilter
	}

	return nil
}
