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

// func (a *AndClause) parseClause(data map[string]interface{}) {
// 	for _, childClause := range a.ChildClauses {
// 		switch childClause.(AndClause).Prefix {
// 		case "$and":
// 			newChildClaud := AndClause{
// 				Prefix:       childClause.(AndClause).Prefix,
// 				ChildClauses: childClause.(AndClause).ChildClauses,
// 			}
// 			newChildClaud.parseClause(data)
// 			if !newChildClaud.Result {
// 				a.Result = false
// 				break
// 			}
// 			if allTrue := true; allTrue {
// 				for _, childClause := range a.ChildClauses {
// 					if !childClause.((AndClause)).Result {
// 						allTrue = false
// 						break
// 					}
// 				}
// 				if allTrue {
// 					a.Result = true
// 				}
// 			}
// 		case "$or":
// 			newChildClause := OrClause{
// 				Prefix:       childClause.(OrClause).Prefix,
// 				ChildClauses: childClause.(OrClause).ChildClauses,
// 			}
// 			newChildClause.parseClause(data)
// 			if !newChildClause.Result {
// 				a.Result = false
// 				break
// 			}
// 			a.Result = false
// 		case "$filter":
// 			newChildClause := FilterClause{
// 				Prefix: childClause.(FilterClause).Prefix,
// 				Filter: childClause.(FilterClause).Filter,
// 			}
// 			newChildClause.parseClause(data)
// 			if !newChildClause.Result {
// 				a.Result = false
// 				break
// 			}
// 		}
// 	}
// }

func (c *Clause) parseClause(data map[string]interface{}) {
	for i, childClause := range c.ChildClauses {

		switch childClause.Prefix {
		case "$and":
			newChildClaud := Clause{
				Prefix:       childClause.Prefix,
				ChildClauses: childClause.ChildClauses,
			}
			newChildClaud.parseClause(data)

			// fmt.Printf("---newChildClaud %#v\n", newChildClaud)
			if newChildClaud.Result {
				c.ChildClauses[i].Result = true
				// break
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
			fmt.Println("SECOND--OR")
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
			fmt.Println("SECOND--FILTER")
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

// func (o *OrClause) parseClause(data map[string]interface{}) {
// 	for _, childClause := range o.ChildClauses {
// 		switch childClause.(OrClause).Prefix {
// 		case "$and":
// 			newChildClaud := AndClause{
// 				Prefix:       childClause.(AndClause).Prefix,
// 				ChildClauses: childClause.(AndClause).ChildClauses,
// 			}
// 			newChildClaud.parseClause(data)
// 			if !newChildClaud.Result {
// 				o.Result = false
// 				break
// 			}
// 		case "$or":
// 			newChildClaud := OrClause{
// 				Prefix:       childClause.(OrClause).Prefix,
// 				ChildClauses: childClause.(OrClause).ChildClauses,
// 			}
// 			newChildClaud.parseClause(data)
// 			if !newChildClaud.Result {
// 				o.Result = false
// 				break
// 			}
// 			o.Result = false
// 		case "$filter":
// 			newChildClause := FilterClause{
// 				Prefix: childClause.(FilterClause).Prefix,
// 				Filter: childClause.(FilterClause).Filter,
// 			}
// 			newChildClause.parseClause(data)
// 			if !newChildClause.Result {
// 				o.Result = false
// 				break
// 			}
// 		}
// 	}
// }

// func (f *FilterClause) parseClause(data map[string]interface{}) {
// 	filterClause := FilterClause{
// 		Prefix: f.Prefix,
// 		Filter: f.Filter,
// 	}
// 	filterFunc, ok := _filterFunctions[f.Filter.Category][f.Filter.FuncName]
// 	if !ok {
// 		panic("Filter function is not yet defined")
// 	}
// 	if _, ok := data[filterClause.Filter.Key]; ok {
// 		f.Result = filterFunc(data, f.Filter.Key, f.Filter.Parameter)
// 	} else {
// 		f.Result = true
// 	}
// }

// func parseClause(clause interface{}, data map[string]interface{}) {
// 	switch clause.(ClauseWithResult).Prefix {
// 	case "$filter":
// 		filterClause := clause.(FilterClause)
// 		filterFunc, ok := FilterFunctions[filterClause.Filter.Category][filterClause.Filter.FuncName]
// 		if !ok {
// 			panic("Filter function is not yet defined")
// 		}
// 		if _, ok := data[filterClause.filter.key]; ok {
// 			clause.(ClauseWithResult).Result = filterFunc(data, filterClause.Filter.Key, filterClause.Filter.Parameter)
// 		} else {
// 			clause.(ClauseWithResult).Result = true
// 		}
// 	case "$and":
// 		andClause := clause.(AndClause)
// 		for _, childClause := range andClause.ChildClauses {
// 			parseClause(childClause, data)
// 			if childClause.Result == false {
// 				clause.(ClauseWithResult).Result = false
// 				break
// 			}
// 		}
// 		if allTrue := true; allTrue {
// 			for _, childClause := range andClause.ChildClauses {
// 				if childClause.Result != true {
// 					allTrue = false
// 					break
// 				}
// 			}
// 			if allTrue {
// 				clause.(ClauseWithResult).Result = true
// 			}
// 		}
// 	case "$or":
// 		orClause := clause.OrClause
// 		for _, childClause := range orClause.ChildClauses {
// 			parseClause(childClause, data)
// 			if childClause.result == true {
// 				clause.(ClauseWithResult).Result = true
// 				break
// 			}
// 		}
// 		clause.(ClauseWithResult).Result = false
// 	}
// }

// func structFieldExists(obj interface{}, field string) bool {
// 	val := reflect.ValueOf(obj)
// 	if val.Kind() == reflect.Ptr {
// 		val = val.Elem()
// 	}
// 	if val.Kind() != reflect.Struct {
// 		return false
// 	}
// 	_, exists := val.Type().FieldByName(field)
// 	return exists
// }

// type MyStruct struct {
//     Field1 string
//     Field2 int
// }

// s := MyStruct{Field1: "hello", Field2: 42}

// exists := structFieldExists(s, "Field1")
// fmt.Println(exists)  // true

// exists = structFieldExists(s, "Field3")
// fmt.Println(exists)  // false

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
	// fmt.Printf("\n-------- RESULT %#v\n\n", result)
	return result
}

func ProcessQuery(filters interface{}, item map[string]interface{}) bool {
	fmt.Printf("\n-------- item %#v\n\n", item)
	var result bool

	switch filters.(Clause).Prefix {
	case "$and":
		andFilter := Clause{
			Prefix:       filters.(Clause).Prefix,
			ChildClauses: filters.(Clause).ChildClauses,
		}
		andFilter.parseClause(item)
		result = andFilter.Result
		// fmt.Printf("andFilter.Result %#v\n", andFilter.Result)
		if allTrue := true; allTrue {
			for _, childClause := range filters.(Clause).ChildClauses {
				// fmt.Printf("---childClause %#v\n", childClause)
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
			fmt.Printf("\n-------- OR %#v\n", childClause)
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
		// var subclauses []interface{}
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).And {
			theClause := DeserialiseQuery(subclause)
			// subclauses = append(subclauses, DeserialiseQuery(subclause))
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$and",
			ChildClauses: subClauses,
		}
	}
	if query.(FilterClauses).Or != nil {
		// var subclauses []interface{}
		var subClauses []Clause
		for _, subclause := range query.(FilterClauses).Or {
			theClause := DeserialiseQuery(subclause)
			// subclauses = append(subclauses, DeserialiseQuery(subclause))
			subClauses = append(subClauses, theClause.(Clause))
		}
		return Clause{
			Prefix:       "$or",
			ChildClauses: subClauses}
	}
	panic("Expected JSON to contain $filter, $or or $and")
}
