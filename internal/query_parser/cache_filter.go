package query_parser

type FilterCategoryType string

const (
	Number   FilterCategoryType = "number"
	String   FilterCategoryType = "string"
	Date     FilterCategoryType = "date"
	Boolean  FilterCategoryType = "boolean"
	Array    FilterCategoryType = "array"
	Duration FilterCategoryType = "duration"
	Enum     FilterCategoryType = "enum"
	Tag      FilterCategoryType = "tag"
)
