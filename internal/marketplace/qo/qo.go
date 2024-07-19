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
	FilterTypeSlug QueryOptionsFilterType = 7
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

func WithCriteria(criteria QueryOptionCriteria) QueryOption {
	return func(o *QueryOptions) {
		o.Criteria = append(o.Criteria, criteria)
	}
}

func WithSlug(slug string) QueryOption {
	return WithCriteria(QueryOptionCriteria{
		FilterType: FilterTypeSlug,
		Value:      slug,
	})
}
