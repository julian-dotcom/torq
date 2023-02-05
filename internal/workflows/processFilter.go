package workflows

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/tags"
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

type FilterFunc func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool

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
			filterFunc, ok := GetFilterFunctions()[filterClause.Filter.Category][filterClause.Filter.FuncName]
			if !ok {
				panic("Filter function is not yet defined")
			}
			if _, ok := data[strings.ToLower(filterClause.Filter.Key)]; ok {
				c.ChildClauses[i].Result = filterFunc(data, strings.ToLower(filterClause.Filter.Key), filterClause.Filter.Parameter)
			} else {
				c.ChildClauses[i].Result = true
			}
		}
	}
}

func GetFilterFunctions() map[FilterCategoryType]map[string]FilterFunc {
	return map[FilterCategoryType]map[string]FilterFunc{
		FilterCategoryTypeNumber: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat == filterValueFloat
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat != filterValueFloat
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat > filterValueFloat
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat >= filterValueFloat
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat < filterValueFloat
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat <= filterValueFloat
			},
		},
		FilterCategoryTypeDuration: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat == filterValueFloat
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat != filterValueFloat
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat > filterValueFloat
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat >= filterValueFloat
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat < filterValueFloat
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDuration) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat <= filterValueFloat
			},
		},
		FilterCategoryTypeString: {
			"like": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
					return false
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
			"notLike": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
					return false
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return !strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
		},
		FilterCategoryTypeEnum: {
			"like": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: dataValueString) so defaulting to false instead of a panic!")
					return false
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
			"notLike": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueString, ok := dataMap[dataKey].(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: dataValueString) so defaulting to false instead of a panic!")
					return false
				}
				filterValueString, ok := filterValue.(string)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueString) so defaulting to false instead of a panic!")
					return false
				}
				return !strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
			},
		},
		FilterCategoryTypeDate: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat == filterValueFloat
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat != filterValueFloat
			},
			"gt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat > filterValueFloat
			},
			"gte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat >= filterValueFloat
			},
			"lt": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat < filterValueFloat
			},
			"lte": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
				if err != nil {
					log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeDate) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueFloat <= filterValueFloat
			},
		},
		FilterCategoryTypeBoolean: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueBoolean, ok := dataMap[dataKey].(bool)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeBoolean: dataValueBoolean) so defaulting to false instead of a panic!")
					return false
				}
				filterValueBoolean, ok := filterValue.(bool)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueBoolean) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueBoolean == filterValueBoolean
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				dataValueBoolean, ok := dataMap[dataKey].(bool)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeBoolean: dataValueBoolean) so defaulting to false instead of a panic!")
					return false
				}
				filterValueBoolean, ok := filterValue.(bool)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeEnum: filterValueBoolean) so defaulting to false instead of a panic!")
					return false
				}
				return dataValueBoolean != filterValueBoolean
			},
		},
		FilterCategoryTypeArray: {
			"eq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				// TODO FIXME this will not work and panic???
				dataValueArray, ok := dataMap[dataKey].([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: dataValueArray) so defaulting to false!")
					return false
				}
				filterValueArray, ok := filterValue.([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: filterValueArray) so defaulting to false!")
					return false
				}
				for _, dataValueEntry := range dataValueArray {
					foundIt := false
					for _, filterValueEntry := range filterValueArray {
						if dataValueEntry == filterValueEntry {
							foundIt = true
						}
					}
					if !foundIt {
						return false
					}
				}
				return true
			},
			"neq": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				// TODO FIXME this will not work and panic???
				dataValueArray, ok := dataMap[dataKey].([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: dataValueArray) so defaulting to false!")
					return false
				}
				filterValueArray, ok := filterValue.([]interface{})
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeArray: filterValueArray) so defaulting to false!")
					return false
				}
				for _, dataValueEntry := range dataValueArray {
					foundIt := false
					for _, filterValueEntry := range filterValueArray {
						if dataValueEntry == filterValueEntry {
							foundIt = true
						}
					}
					if !foundIt {
						return true
					}
				}
				return false
			},
		},
		FilterCategoryTypeTag: {
			"any": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if dataMap[dataKey] == nil {
					return false
				}
				dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag: dataValueTags) so defaulting to false!")
					return false
				}
				tagResponses, tagResponsesOk := dataMap[dataKey].([]tags.TagResponse)
				if tagResponsesOk {
					for _, tag := range tagResponses {
						for _, dataValueTag := range dataValueTags {
							if tag.TagId == dataValueTag.TagId {
								return true
							}
						}
					}
				}
				tgs, tagsOk := dataMap[dataKey].([]tags.Tag)
				if tagsOk {
					for _, tag := range tgs {
						for _, dataValueTag := range dataValueTags {
							if tag.TagId == dataValueTag.TagId {
								return true
							}
						}
					}
				}
				if !tagsOk && !tagResponsesOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag) so defaulting to false!")
					return false
				}
				return false
			},
			"notAny": func(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
				if dataMap[dataKey] == nil {
					return true
				}
				dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
				if !ok {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag: dataValueTags) so defaulting to false!")
					return false
				}
				tagResponses, tagsOk := dataMap[dataKey].([]tags.TagResponse)
				if tagsOk {
					for _, tag := range tagResponses {
						for _, dataValueTag := range dataValueTags {
							if tag.TagId == dataValueTag.TagId {
								return false
							}
						}
					}
				}
				tgs, tagResponsesOk := dataMap[dataKey].([]tags.Tag)
				if tagResponsesOk {
					for _, tag := range tgs {
						for _, dataValueTag := range dataValueTags {
							if tag.TagId == dataValueTag.TagId {
								return false
							}
						}
					}
				}
				if !tagsOk && !tagResponsesOk {
					log.Error().Msgf("could not run the filter function (FilterCategoryTypeTag) so defaulting to false!")
					return false
				}
				return true
			},
		},
	}
}

func getFloats(dataValueOfUnknownType interface{}, filterValueOfUnknownType interface{}) (float64, float64, error) {
	dataValueFloat, err := getFloat(dataValueOfUnknownType)
	if err != nil {
		return math.NaN(), math.NaN(), err
	}
	filterValueFloat, err := getFloat(filterValueOfUnknownType)
	if err != nil {
		return math.NaN(), math.NaN(), err
	}
	return dataValueFloat, filterValueFloat, nil
}

// Note that large int64's (over 2 pow 53) will be rounded when converted to float64,
// so if you're thinking of float64 as a "universal" number type, take its limitations into account.
func getInt(unknownType interface{}) (int64, error) {
	switch i := unknownType.(type) {
	case float64:
		return int64(i), nil
	case float32:
		return int64(i), nil
	case int64:
		return i, nil
	case int32:
		return int64(i), nil
	case int:
		return int64(i), nil
	case uint64:
		return int64(i), nil
	case uint32:
		return int64(i), nil
	case uint:
		return int64(i), nil
	case string:
		parameter, err := strconv.ParseInt(i, 10, 64)
		return parameter, err
	default:
		var floatType = reflect.TypeOf(int64(0))
		var stringType = reflect.TypeOf("")
		v := reflect.ValueOf(unknownType)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Int(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			parameter, err := strconv.ParseInt(s, 10, 64)
			return parameter, err
		} else {
			return 0, fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
}

// Note that large int64's (over 2 pow 53) will be rounded when converted to float64,
// so if you're thinking of float64 as a "universal" number type, take its limitations into account.
func getFloat(unknownType interface{}) (float64, error) {
	switch i := unknownType.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		parameter, err := strconv.ParseFloat(i, 64)
		return parameter, err
	default:
		var floatType = reflect.TypeOf(float64(0))
		var stringType = reflect.TypeOf("")
		v := reflect.ValueOf(unknownType)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			parameter, err := strconv.ParseFloat(s, 64)
			return parameter, err
		} else {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
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
