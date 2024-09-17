package qo

type QueryOptions struct {
	PageNumber int                   `json:"pageNumber"`
	PageSize   int                   `json:"pageSize"`
	SortBy     QueryOptionSortBy     `json:"sortBy"`
	Direction  QueryOptionDirection  `json:"direction"`
	Criteria   []QueryOptionCriteria `json:"criteria"`
}

type QueryOption func(o *QueryOptions)

type QueryOptionSortBy int

const (
	SortByLastUpdated   QueryOptionSortBy = 1
	SortByInstalls      QueryOptionSortBy = 4
	SortByPublishedDate QueryOptionSortBy = 10
)

type QueryOptionDirection int

const (
	DirectionAsc QueryOptionDirection = 1
	DirectionDes QueryOptionDirection = 2
)

type QueryOptionCriteria struct {
	FilterType QueryOptionsFilterType `json:"filterType"`
	Value      string                 `json:"value"`
}

type QueryOptionsFilterType int

const (
	FilterTypeCategory  QueryOptionsFilterType = 5
	FilterTypeSlug      QueryOptionsFilterType = 7
	FilterTypeUnknown8  QueryOptionsFilterType = 8
	FilterTypeUnknown10 QueryOptionsFilterType = 10
	FilterTypeUnknown12 QueryOptionsFilterType = 12
)

func WithPageNumber(pageNumber int) QueryOption {
	return func(o *QueryOptions) {
		o.PageNumber = pageNumber
	}
}

func WithPageSize(pageSize int) QueryOption {
	return func(o *QueryOptions) {
		o.PageSize = pageSize
	}
}

func WithSortBy(sortBy QueryOptionSortBy) QueryOption {
	return func(o *QueryOptions) {
		o.SortBy = sortBy
	}
}

func WithDirection(direction QueryOptionDirection) QueryOption {
	return func(o *QueryOptions) {
		o.Direction = direction
	}
}

func WithCriteria(filterType QueryOptionsFilterType, value string) QueryOption {
	return func(o *QueryOptions) {
		o.Criteria = append(o.Criteria, QueryOptionCriteria{
			FilterType: filterType,
			Value:      value,
		})
	}
}

func WithSlug(slug string) QueryOption {
	return WithCriteria(FilterTypeSlug, slug)
}
