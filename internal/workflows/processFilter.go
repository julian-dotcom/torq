package workflows

import "strings"

type FilterCategoryType string

const (
	FilterCategoryTypeNumber   FilterCategoryType = "number"
	FilterCategoryTypeString   FilterCategoryType = "string"
	FilterCategoryTypeDate     FilterCategoryType = "date"
	FilterCategoryTypeBoolean  FilterCategoryType = "boolean"
	FilterCategoryTypeArray    FilterCategoryType = "array"
	FilterCategoryTypeDuration FilterCategoryType = "duration"
	FilterCategoryTypeEnum     FilterCategoryType = "enum"
	FilterCategoryTypeTag      FilterCategoryType = "tag"
)

type FilterParameterType interface{}

type FilterFunc func(input map[string]interface{}, key string, parameter FilterParameterType) bool

type SelectOptionType struct {
	value string
	label string
}

type FilterInterface struct {
	Category      FilterCategoryType
	FuncName      string
	Parameter     FilterParameterType
	Key           string
	SelectOptions []SelectOptionType
	Value         interface{}
	Label         string
}

type FilterClause struct {
	Prefix string
	Filter FilterInterface
}

type AndClause struct {
	Prefix       string
	ChildClauses []Clause
}

func (f *FilterClause) ToJSON() map[string]interface{} {
	return map[string]interface{}{f.Prefix: f.Filter}
}

func (a *AndClause) AddChildClause(clause Clause) {
	a.ChildClauses = append(a.ChildClauses, clause)
}

func (a *AndClause) ToJSON() map[string]interface{} {
	childJSON := make([]map[string]interface{}, len(a.ChildClauses))
	for i, child := range a.ChildClauses {
		childJSON[i] = child.ToJSON()
	}
	return map[string]interface{}{a.Prefix: childJSON}
}

type OrClause struct {
	Prefix       string
	ChildClauses []Clause
}

type Clause interface {
	Length() int
	ToJSON() map[string]interface{}
}

type ClauseWithResult struct {
	Prefix string
	Result bool
}

func parseClause(clause interface{}, data interface{}) {
	switch clause.(ClauseWithResult).Prefix {
	case "$filter":
		filterClause := clause.(FilterClause)
		filterFunc, ok := FilterFunctions[filterClause.Filter.Category][filterClause.Filter.FuncName]
		if !ok {
			panic("Filter function is not yet defined")
		}
		if _, ok := data[filterClause.filter.key]; ok {
			clause.(ClauseWithResult).Result = filterFunc(data, filterClause.Filter.Key, filterClause.Filter.Parameter)
		} else {
			clause.(ClauseWithResult).Result = true
		}
	case "$and":
		andClause := clause.(AndClause)
		for _, childClause := range andClause.ChildClauses {
			parseClause(childClause, data)
			if childClause.Result == false {
				clause.(ClauseWithResult).Result = false
				break
			}
		}
		if allTrue := true; allTrue {
			for _, childClause := range andClause.ChildClauses {
				if childClause.Result != true {
					allTrue = false
					break
				}
			}
			if allTrue {
				clause.(ClauseWithResult).Result = true
			}
		}
	case "$or":
		orClause := clause.OrClause
		for _, childClause := range orClause.ChildClauses {
			parseClause(childClause, data)
			if childClause.result == true {
				clause.(ClauseWithResult).Result = true
				break
			}
		}
		clause.(ClauseWithResult).Result = false
	}
}

var FilterFunctions = map[FilterCategoryType]map[string]FilterFunc{
	FilterCategoryTypeNumber: {
		"eq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] == parameter
		},
		"neq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] != parameter
		},
		"gt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) > parameter.(float64)
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) >= parameter.(float64)
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) < parameter.(float64)
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) <= parameter.(float64)
		},
	},
	FilterCategoryTypeDuration: {
		"eq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] == parameter
		},
		"neq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] != parameter
		},
		"gt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) > parameter.(float64)
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) >= parameter.(float64)
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) < parameter.(float64)
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) <= parameter.(float64)
		},
	},
	FilterCategoryTypeString: {
		"like": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return strings.Contains(strings.ToLower(input[key].(string)), strings.ToLower(parameter.(string)))
		},
		"notLike": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return !strings.Contains(strings.ToLower(input[key].(string)), strings.ToLower(parameter.(string)))
		},
	},
	FilterCategoryTypeEnum: {
		"like": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return strings.Contains(strings.ToLower(input[key].(string)), strings.ToLower(parameter.(string)))
		},
		"notLike": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return !strings.Contains(strings.ToLower(input[key].(string)), strings.ToLower(parameter.(string)))
		},
	},
	FilterCategoryTypeDate: {
		"eq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] == parameter
		},
		"neq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key] != parameter
		},
		"gt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) > parameter.(float64)
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) >= parameter.(float64)
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) < parameter.(float64)
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(float64) <= parameter.(float64)
		},
	},
	FilterCategoryTypeBoolean: {
		"eq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(bool) == parameter
		},
		"neq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(bool) != parameter
		},
	},
	FilterCategoryTypeArray: {
		"eq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			paramArray := parameter.([]interface{})
			for _, value := range input[key].([]interface{}) {
				for _, paramValue := range paramArray {
					if value == paramValue {
						return true
					}
				}
			}
			return false
		},
		"neq": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			paramArray := parameter.([]interface{})
			for _, value := range input[key].([]interface{}) {
				for _, paramValue := range paramArray {
					if value == paramValue {
						return false
					}
				}
			}
			return true
		},
	},
}

func applyFilters(filters Clause, data []interface{}) []interface{} {
	var result []interface{}
	for _, item := range data {
		if processQuery(filters, item) {
			result = append(result, item)
		}
	}
	return result
}

func processQuery(filters Clause, item interface{}) bool {
	// TODO to be done
	parseClause(filters, item)
	return true
}
