package workflows

import (
	"github.com/lncapital/torq/internal/tags"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func filterCategoryEnumAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	dataValue, ok := dataMap[dataKey].(string)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryArray: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		if filterValueItem == dataValue {
			return true
		}
	}
	return false
}

func filterCategoryEnumNotAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	dataValueArray, ok := dataMap[dataKey].(string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryEnumNotAny: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryEnumNotAny: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		if filterValueItem == dataValueArray {
			return false
		}
	}
	return true
}

func filterCategoryArrayAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	dataValueArray, ok := dataMap[dataKey].([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryArrayAny: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryArrayAny: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		for _, dataValueItem := range dataValueArray {
			if filterValueItem == dataValueItem {
				return true
			}
		}
	}
	return false
}

func filterCategoryArrayNotAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	dataValueArray, ok := dataMap[dataKey].([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryArrayNotAny: dataValueArray) so defaulting to false!")
		return false
	}
	filterValueArray, ok := filterValue.([]string)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryArrayNotAny: filterValueArray) so defaulting to false!")
		return false
	}
	for _, filterValueItem := range filterValueArray {
		for _, dataValueItem := range dataValueArray {
			if filterValueItem == dataValueItem {
				return false
			}
		}
	}
	return true
}

func filterCategoryTypeTagAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeTagAny: dataValueTags) so defaulting to false!")
		return false
	}
	filterValueTagResponses, tagResponsesOk := filterValue.([]tags.TagResponse)
	if tagResponsesOk {
		for _, tag := range filterValueTagResponses {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	filterValueTags, tagsOk := filterValue.([]tags.Tag)
	if tagsOk {
		for _, tag := range filterValueTags {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	filterValueTagIdsO, tagIdsOk := filterValue.([]interface{})
	if tagIdsOk {
	filterLoop:
		for _, tagIdO := range filterValueTagIdsO {
			for _, dataValueTag := range dataValueTags {
				tagId, err := getFloat(tagIdO)
				if err != nil {
					log.Error().Msgf("could not run convert interface into tagId in filter function (filterCategoryTypeTagAny)")
					tagIdsOk = false
					break filterLoop
				}
				if int(tagId) == dataValueTag.TagId {
					return true
				}
			}
		}
	}
	if !tagsOk && !tagResponsesOk && !tagIdsOk {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeTagAny) so defaulting to false!")
		return false
	}
	return false
}

func filterCategoryTypeTagNotAny(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {

	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	// If the filter value is nil, returns false.
	if filterValue == nil {
		return false
	}
	// Converts the data value to a slice of tags.Tag, and if it fails, returns false.
	dataValueTags, ok := dataMap[dataKey].([]tags.Tag)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeTagNotAny: dataValueTags) so defaulting to false!")
		return false
	}

	// If the filter value is a slice of tags.TagResponse, iterates over the filter value tags and the data value tags
	//   and returns false if any of the filter tags match any of the data value tags.
	filterValueTagResponses, tagResponsesOk := filterValue.([]tags.TagResponse)
	if tagResponsesOk {
		for _, tag := range filterValueTagResponses {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// If the filter value is a slice of tags.Tag, iterates over the filter value tags and the data value tags
	//   and returns false if any of the filter tags match any of the data value tags.
	filterValueTags, tagsOk := filterValue.([]tags.Tag)
	if tagsOk {
		for _, tag := range filterValueTags {
			for _, dataValueTag := range dataValueTags {
				if tag.TagId == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// If the filter value is a slice of interfaces, iterates over the filter value tag ids and the data value tags
	//   and returns false if any of the filter tag ids match any of the data value tags.
	filterValueTagIdsO, tagIdsOk := filterValue.([]interface{})
	if tagIdsOk {
	filterLoop:
		for _, tagIdO := range filterValueTagIdsO {
			for _, dataValueTag := range dataValueTags {
				tagId, err := getFloat(tagIdO)
				if err != nil {
					log.Error().Msgf("could not run convert interface into tagId in filter function (FilterCategoryTypeTag)")
					tagIdsOk = false
					break filterLoop
				}
				if int(tagId) == dataValueTag.TagId {
					return false
				}
			}
		}
	}

	// Converts the tag id to a float and returns false if it fails.
	if !tagsOk && !tagResponsesOk && !tagIdsOk {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeTagNotAny) so defaulting to false!")
		return false
	}

	return true
}

func filterCategoryTypeBooleanEq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if filterValue == nil && dataMap[dataKey] == nil {
		return true
	}
	dataValueBoolean, ok := dataMap[dataKey].(bool)
	if !ok {
		dataValueBooleanPointer, ok := dataMap[dataKey].(*bool)
		if !ok {
			log.Error().Msgf("could not run the filter function (filterCategoryTypeBooleanEq: dataValueBoolean) so defaulting to false instead of a panic!")
			return false
		}
		dataValueBoolean = *dataValueBooleanPointer
	}
	filterValueBoolean, ok := filterValue.(bool)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeBooleanEq: filterValueBoolean) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueBoolean == filterValueBoolean
}

func filterCategoryTypeBooleanNeq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if filterValue == nil && dataMap[dataKey] == nil {
		return false
	}
	if filterValue != nil && dataMap[dataKey] == nil {
		return true
	}
	if filterValue == nil && dataMap[dataKey] != nil {
		return true
	}
	dataValueBoolean, ok := dataMap[dataKey].(bool)
	if !ok {
		dataValueBooleanPointer, ok := dataMap[dataKey].(*bool)
		if !ok {
			log.Error().Msgf("could not run the filter function (filterCategoryTypeBooleanNeq: dataValueBoolean) so defaulting to false instead of a panic!")
			return false
		}
		dataValueBoolean = *dataValueBooleanPointer
	}
	filterValueBoolean, ok := filterValue.(bool)
	if !ok {
		log.Error().Msgf("could not run the filter function (filterCategoryTypeBooleanNeq: filterValueBoolean) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueBoolean != filterValueBoolean
}

func filterCategoryTypeNumberEq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat == filterValueFloat
}

func filterCategoryTypeNumberNeq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil {
		return false
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat != filterValueFloat
}

func filterCategoryTypeNumberGt(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return false
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat > filterValueFloat
}

func filterCategoryTypeNumberGte(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat >= filterValueFloat
}

func filterCategoryTypeNumberLt(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return false
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat < filterValueFloat
}

func filterCategoryTypeNumberLte(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueFloat, filterValueFloat, err := getFloats(dataMap[dataKey], filterValue)
	if err != nil {
		log.Error().Err(err).Msgf("could not run the filter function (FilterCategoryTypeNumber) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueFloat <= filterValueFloat
}

func filterCategoryTypeStringLike(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueString, ok := dataMap[dataKey].(string)
	if !ok {
		dataValueStringPointer, ok := dataMap[dataKey].(*string)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
			return false
		}
		dataValueString = *dataValueStringPointer
	}
	filterValueString, ok := filterValue.(string)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
		return false
	}
	return strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
}

func filterCategoryTypeStringNotLike(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil {
		return false
	}
	dataValueString, ok := dataMap[dataKey].(string)
	if !ok {
		dataValueStringPointer, ok := dataMap[dataKey].(*string)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: dataValueString) so defaulting to false instead of a panic!")
			return false
		}
		dataValueString = *dataValueStringPointer
	}
	filterValueString, ok := filterValue.(string)
	if !ok {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeString: filterValueString) so defaulting to false instead of a panic!")
		return false
	}
	return !strings.Contains(strings.ToLower(dataValueString), strings.ToLower(filterValueString))
}

func filterCategoryTypeDateEq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueTime.Equal(filterValueTime)
}

func filterCategoryTypeDateNeq(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return true
	}
	if filterValue == nil {
		return false
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return !dataValueTime.Equal(filterValueTime)
}

func filterCategoryTypeDateGt(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return false
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueTime.After(filterValueTime)
}

func filterCategoryTypeDateGte(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueTime.After(filterValueTime) || dataValueTime.Equal(filterValueTime)
}

func filterCategoryTypeDateLt(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return false
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueTime.Before(filterValueTime)
}

func filterCategoryTypeDateLte(dataMap map[string]interface{}, dataKey string, filterValue FilterParameterType) bool {
	if isNil(dataMap[dataKey]) != (filterValue == nil) {
		return false
	}
	if filterValue == nil {
		return true
	}
	dataValueTime, ok := dataMap[dataKey].(time.Time)
	if !ok {
		dataValueTimePointer, ok := dataMap[dataKey].(*time.Time)
		if !ok {
			log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: dataValueTime) so defaulting to false instead of a panic!")
			return false
		}
		dataValueTime = *dataValueTimePointer
	}
	dataValueTime = truncateToMinute(dataValueTime)
	filterValueTime, timeOk := filterValue.(time.Time)
	if timeOk {
		filterValueTime = truncateToMinute(filterValueTime)
	}
	filterValueString, stringOk := filterValue.(string)
	if stringOk {
		var err error
		filterValueTime, err = time.Parse("2006-01-02T15:04", filterValueString)
		if err != nil {
			log.Error().Msgf("could not parse the filter function criteria (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
			return false
		}
	}
	if !timeOk && !stringOk {
		log.Error().Msgf("could not run the filter function (FilterCategoryTypeDate: filterValueTime) so defaulting to false instead of a panic!")
		return false
	}
	return dataValueTime.Before(filterValueTime) || dataValueTime.Equal(filterValueTime)
}
