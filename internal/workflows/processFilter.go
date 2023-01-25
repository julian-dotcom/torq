package workflows

import (
	"fmt"
	"strings"
)

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

type FilterClauses struct {
	And    []FilterClauses `json:"$and"`
	Or     []FilterClauses `json:"$or"`
	Filter Filter          `json:"$filter"`
}
type Parameter string

type Filter struct {
	FuncName  string      `json:"funcName"`
	Key       string      `json:"key"`
	Parameter interface{} `json:"parameter"`
	Category  string      `json:"category"`
}

type FilterParameterType interface{}

type FilterFunc func(input map[string]interface{}, key string, parameter FilterParameterType) bool

type SelectOptionType struct {
	Value string
	Label string
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
	Result bool
}

type AndClause struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

func (f *FilterClause) ToJSON() map[string]interface{} {
	return map[string]interface{}{f.Prefix: f.Filter}
}

func (a *AndClause) AddChildClause(clause Clause) {
	a.ChildClauses = append(a.ChildClauses, clause)
}

type OrClause struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

type Clause struct {
	Prefix       string
	ChildClauses []Clause
	Filter       FilterInterface
	Result       bool
}

type ClauseWithResult struct {
	Prefix       string
	ChildClauses []interface{}
	Result       bool
}

func (c *Clause) parseClause(data map[string]interface{}) {
	for i, childClause := range c.ChildClauses {

		switch childClause.Prefix {
		case "$and":
			newChildClaud := Clause{
				Prefix:       childClause.Prefix,
				ChildClauses: childClause.ChildClauses,
			}
			newChildClaud.parseClause(data)

			if newChildClaud.Result {
				c.ChildClauses[i].Result = true
			}

			if allTrue := true; allTrue {
				for _, subChildClause := range newChildClaud.ChildClauses {
					if !subChildClause.Result {
						allTrue = false
						break
					}
				}
				if allTrue {
					c.Result = true
				}
			}
		case "$or":
			newChildClaud := Clause{
				Prefix:       childClause.Prefix,
				ChildClauses: childClause.ChildClauses,
			}
			newChildClaud.parseClause(data)
			for _, subChildClause := range newChildClaud.ChildClauses {
				if subChildClause.Result {
					c.ChildClauses[i].Result = true
				}
			}
		case "$filter":
			filterClause := FilterClause{
				Prefix: childClause.Prefix,
				Filter: childClause.Filter,
			}
			filterFunc, ok := FilterFunctions[filterClause.Filter.Category][filterClause.Filter.FuncName]
			if !ok {
				panic("Filter function is not yet defined")
			}
			if _, ok := data[filterClause.Filter.Key]; ok {
				c.ChildClauses[i].Result = filterFunc(data, filterClause.Filter.Key, filterClause.Filter.Parameter)
			} else {
				c.ChildClauses[i].Result = true
			}
		}
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
			fmt.Println("input[key]", input[key].(int64))
			fmt.Println("parameter", int64(parameter.(float64)))
			fmt.Println("SOO", input[key].(int64) > int64(parameter.(float64)))
			return input[key].(int64) > int64(parameter.(float64))
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) >= int64(parameter.(float64))
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) < int64(parameter.(float64))
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) <= int64(parameter.(float64))
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
			return input[key].(int64) > int64(parameter.(float64))
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) >= int64(parameter.(float64))
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) < int64(parameter.(float64))
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) <= int64(parameter.(float64))
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
			return input[key].(int64) > int64(parameter.(float64))
		},
		"gte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) >= int64(parameter.(float64))
		},
		"lt": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) < int64(parameter.(float64))
		},
		"lte": func(input map[string]interface{}, key string, parameter FilterParameterType) bool {
			return input[key].(int64) <= int64(parameter.(float64))
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

func ApplyFilters(filters interface{}, data []map[string]interface{}) []interface{} {

	var result []interface{}
	for _, item := range data {
		if ProcessQuery(DeserialiseQuery(filters), item) {
			result = append(result, item)
		}
	}
	return result
}

func ProcessQuery(filters interface{}, item map[string]interface{}) bool {
	var result bool

	switch filters.(Clause).Prefix {
	case "$and":
		andFilter := Clause{
			Prefix:       filters.(Clause).Prefix,
			ChildClauses: filters.(Clause).ChildClauses,
		}
		andFilter.parseClause(item)
		result = andFilter.Result
		if allTrue := true; allTrue {
			for _, childClause := range filters.(Clause).ChildClauses {
				if !childClause.Result {
					allTrue = false
					break
				}
			}
			if allTrue {
				result = true
			}
		}
	case "$or":
		orFilter := Clause{
			Prefix:       filters.(Clause).Prefix,
			ChildClauses: filters.(Clause).ChildClauses,
		}
		orFilter.parseClause(item)
		result = orFilter.Result
		for _, childClause := range filters.(Clause).ChildClauses {
			if childClause.Result {
				result = childClause.Result
				break
			}
		}
	case "$filter":
		filterClause := Clause{
			Prefix: filters.(Clause).Prefix,
			Filter: filters.(Clause).Filter,
		}
		filterClause.parseClause(item)
		result = filterClause.Result
	}
	return result
}

func DeserialiseQuery(query interface{}) interface{} {
	if query == nil {
		return Clause{}
	}
	if query.(FilterClauses).Filter.FuncName != "" {

		filter := FilterInterface{
			Parameter: query.(FilterClauses).Filter.Parameter,
			FuncName:  query.(FilterClauses).Filter.FuncName,
			Key:       query.(FilterClauses).Filter.Key,
			Category:  FilterCategoryType(query.(FilterClauses).Filter.Category),
		}
		return Clause{
			Prefix: "$filter",
			Filter: filter,
		}
	}
	if query.(FilterClauses).And != nil {
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).And {
			theClause := DeserialiseQuery(subclause)
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$and",
			ChildClauses: subClauses,
		}
	}
	if query.(FilterClauses).Or != nil {
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).Or {
			theClause := DeserialiseQuery(subclause)
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$or",
			ChildClauses: subClauses}
	}
	panic("Expected JSON to contain $filter, $or or $and")
}
