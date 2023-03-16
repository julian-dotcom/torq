package workflows

import (
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/testutil"
	"testing"
)

func TestFilterCategoryEnumAny(t *testing.T) {

	dataKey := "key1"
	dataMap := map[string]interface{}{
		dataKey: "hello",
	}
	emptyDataMap := map[string]interface{}{
		dataKey: nil,
	}

	testCases := []struct {
		name        string
		filterValue interface{}
		dataMap     map[string]interface{}
		want        bool
	}{
		{
			name:        "nil filter value",
			filterValue: nil,
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "nil filter value and nil data value",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        true,
		},
		{
			name:        "nil data value non-nil filter",
			filterValue: []string{"hello", "world"},
			dataMap:     nil,
			want:        false,
		},
		{
			name:        "empty filter value and non-empty data value",
			filterValue: nil,
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "overlapping filter with enum",
			filterValue: []string{"world", "hello", "aaaa"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "one matching array item with enum",
			filterValue: []string{"hello"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "Mismatching array item with enum",
			filterValue: []string{"not", "here"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryEnumAny(tc.dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryArrayAny() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryArrayAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterCategoryEnumNotAny(t *testing.T) {

	dataKey := "key1"
	dataMap := map[string]interface{}{
		dataKey: "hello",
	}
	emptyDataMap := map[string]interface{}{
		dataKey: nil,
	}

	testCases := []struct {
		name        string
		filterValue interface{}
		dataMap     map[string]interface{}
		want        bool
	}{
		{
			name:        "nil filter value",
			filterValue: nil,
			dataMap:     nil,
			want:        false,
		},
		{
			name:        "nil filter value and nil data value",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        false,
		},
		{
			name:        "nil data value non-nil filter",
			filterValue: []string{"hello", "world"},
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "empty filter value and non-empty data value",
			filterValue: nil,
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "overlapping filter with enum",
			filterValue: []string{"world", "hello", "aaaa"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "one matching array item with enum",
			filterValue: []string{"hello"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "Mismatching array item with enum",
			filterValue: []string{"not", "here"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryEnumNotAny(tc.dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryArrayNotAny() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryArrayNotAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterCategoryArrayAny(t *testing.T) {

	dataKey := "key1"
	dataMap := map[string]interface{}{
		dataKey: []string{"hello", "world"},
	}
	emptyDataMap := map[string]interface{}{
		dataKey: nil,
	}

	testCases := []struct {
		name        string
		filterValue interface{}
		dataMap     map[string]interface{}
		want        bool
	}{
		{
			name:        "nil filter value",
			filterValue: nil,
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "nil filter value and nil data value",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        true,
		},
		{
			name:        "nil data value non-nil filter",
			filterValue: []string{"hello", "world"},
			dataMap:     nil,
			want:        false,
		},
		{
			name:        "empty filter value and non-empty data value",
			filterValue: nil,
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "overlapping array",
			filterValue: []string{"world", "hello", "aaaa"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "full matching array",
			filterValue: []string{"world", "hello"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "one matching array item",
			filterValue: []string{"hello"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "Missmatching array items",
			filterValue: []string{"not", "here"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryArrayAny(tc.dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryArrayAny() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryArrayAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterCategoryArrayNotAny(t *testing.T) {

	dataKey := "key1"
	dataMap := map[string]interface{}{
		dataKey: []string{"hello", "world"},
	}
	emptyDataMap := map[string]interface{}{
		dataKey: nil,
	}

	testCases := []struct {
		name        string
		filterValue interface{}
		dataMap     map[string]interface{}
		want        bool
	}{
		{
			name:        "nil filter value",
			filterValue: nil,
			dataMap:     nil,
			want:        false,
		},
		{
			name:        "nil filter value and nil data value",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        false,
		},
		{
			name:        "nil data value non-nil filter",
			filterValue: []string{"hello", "world"},
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "empty filter value and non-empty data value",
			filterValue: nil,
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "overlapping array",
			filterValue: []string{"world", "hello", "aaaa"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "full matching array",
			filterValue: []string{"world", "hello"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "one matching array item",
			filterValue: []string{"hello"},
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "Missmatching array items",
			filterValue: []string{"not", "here"},
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryArrayNotAny(tc.dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryArrayNotAny() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryArrayNotAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterCategoryTypeTagAny(t *testing.T) {

	dataKey := "tags"
	dataMap := map[string]interface{}{
		dataKey: []tags.Tag{
			{TagId: 1, Name: "tag1"},
			{TagId: 2, Name: "tag2"},
			{TagId: 3, Name: "tag3"},
		},
	}
	emptyDataMap := map[string]interface{}{
		dataKey: []tags.Tag{},
	}

	testCases := []struct {
		name        string
		filterValue interface{}
		dataMap     map[string]interface{}
		want        bool
	}{
		{
			name:        "nil filter value",
			filterValue: nil,
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "nil filter value and nil data value",
			filterValue: nil,
			dataMap:     dataMap,
			want:        true,
		},
		{
			name:        "nil filter value and no tags",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        true,
		},
		{
			name: "TagResponse filter value with matching tag",
			filterValue: []tags.Tag{
				{TagId: 1, Name: "tag1"},
			},
			dataMap: dataMap,
			want:    true,
		},
		{
			name: "Tag filter value with matching tag",
			filterValue: []tags.Tag{
				{TagId: 3, Name: "tag3"},
			},
			dataMap: dataMap,
			want:    true,
		},
		{
			name: "TagId filter value with matching tag",
			filterValue: []interface{}{
				float64(1),
			},
			dataMap: dataMap,
			want:    true,
		},
		{
			name: "Filter value without matching tag",
			filterValue: []tags.Tag{
				{TagId: 4, Name: "tag4"},
			},
			dataMap: dataMap,
			want:    false,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryTypeTagAny(dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				t.Errorf("filterCategoryTypeTagAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

// Test TestFilterCategoryTypeTagNotAny
func TestFilterCategoryTypeTagNotAny(t *testing.T) {

	dataKey := "tags"
	dataMap := map[string]interface{}{
		dataKey: []tags.Tag{
			{TagId: 1, Name: "tag1"},
			{TagId: 2, Name: "tag2"},
			{TagId: 3, Name: "tag3"},
		},
	}
	emptyDataMap := map[string]interface{}{
		dataKey: []tags.Tag{},
	}

	testCases := []struct {
		name        string
		dataMap     map[string]interface{}
		filterValue interface{}
		want        bool
	}{
		{
			name:        "nil filter value should return false",
			filterValue: nil,
			dataMap:     dataMap,
			want:        false,
		},
		{
			name:        "nil filter value and nil data should return false",
			filterValue: nil,
			dataMap:     emptyDataMap,
			want:        false,
		},
		{
			name: "TagResponse filter value with matching tag",
			filterValue: []tags.TagResponse{
				{
					Tag: tags.Tag{
						TagId: 1,
						Name:  "tag1",
					},
					Channels: nil,
					Nodes:    nil,
				},
			},
			dataMap: dataMap,
			want:    false,
		},
		{
			name: "Tag filter value with matching tag",
			filterValue: []tags.Tag{
				{TagId: 3, Name: "tag3"},
			},
			dataMap: dataMap,
			want:    false,
		},
		{
			name: "TagId filter value with matching tag",
			filterValue: []interface{}{
				float64(1),
			},
			dataMap: dataMap,
			want:    false,
		},
		{
			name: "Filter value without matching tag",
			filterValue: []tags.Tag{
				{TagId: 4, Name: "tag4"},
			},
			dataMap: dataMap,
			want:    true,
		},
		{
			name:        "Invalid filter value",
			filterValue: "invalid",
			dataMap:     dataMap,
			want:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryTypeTagNotAny(tc.dataMap, dataKey, tc.filterValue)
			if got != tc.want {
				t.Errorf("filterCategoryTypeTagNotAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterCategoryTypeBooleanEq(t *testing.T) {

	testCases := []struct {
		name        string
		key         string
		dataMap     map[string]interface{}
		filterValue interface{}
		want        bool
	}{
		{
			name:        "data value is nil and filter is nil",
			key:         "key1",
			filterValue: nil,
			dataMap:     nil,
			want:        true,
		},
		{
			name:        "data value is nil and filter value is nil",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: true,
		},
		{
			name:        "data value is nil filter is false",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: false,
		},
		{
			name:        "data value is nil and data is true",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: false,
		},
		{
			name:        "filter value is nil and data is false",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: false,
		},
		{
			name:        "filter value is nil and data is true",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: false,
		},
		{
			name:        "Boolean filter value is true and data is true",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: true,
		},
		{
			name:        "Boolean filter value is false and data is false",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: true,
		},
		{
			name:        "Boolean filter value is true and data is false",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: false, // filter is true != value is false
		},
		{
			name:        "Boolean filter value is false and data is true",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: false, // filter is false != value is true
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryTypeBooleanEq(tc.dataMap, tc.key, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryTypeBooleanEq() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryTypeBooleanEq() = %v, want %v", got, tc.want)
			}
		})
	}

}

func TestFilterCategoryTypeBooleanNeq(t *testing.T) {

	testCases := []struct {
		name        string
		key         string
		dataMap     map[string]interface{}
		filterValue interface{}
		want        bool
	}{
		{
			name:        "data value is nil and filter is nil",
			key:         "key1",
			filterValue: nil,
			dataMap:     nil,
			want:        false,
		},
		{
			name:        "data value is nil and filter value is nil",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: false,
		},
		{
			name:        "data value is nil filter is false",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: true,
		},
		{
			name:        "data value is nil and data is true",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": nil,
			},
			want: true,
		},
		{
			name:        "filter value is nil and data is false",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: true,
		},
		{
			name:        "filter value is nil and data is true",
			key:         "key1",
			filterValue: nil,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: true,
		},
		{
			name:        "Boolean filter value is true and data is true",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: false,
		},
		{
			name:        "Boolean filter value is false and data is false",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: false,
		},
		{
			name:        "Boolean filter value is true and data is false",
			key:         "key1",
			filterValue: true,
			dataMap: map[string]interface{}{
				"key1": false,
			},
			want: true, // filter is true != value is false
		},
		{
			name:        "Boolean filter value is false and data is true",
			key:         "key1",
			filterValue: false,
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: true, // filter is false != value is true
		},
		{
			name:        "invalid filter",
			key:         "key1",
			filterValue: "invalid",
			dataMap: map[string]interface{}{
				"key1": true,
			},
			want: false, // filter is false != value is true
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCategoryTypeBooleanNeq(tc.dataMap, tc.key, tc.filterValue)
			if got != tc.want {
				testutil.Errorf(t, "filterCategoryTypeBooleanNeq() = %v, want %v", got, tc.want)
			} else {
				testutil.Successf(t, "filterCategoryTypeBooleanNeq() = %v, want %v", got, tc.want)
			}
		})
	}

}

//func TestDeserialiseQuery(t *testing.T) {
//	type args struct {
//		query interface{}
//	}
//	tests := []struct {
//		name string
//		args args
//		want interface{}
//	}{
//		// TODO: Add test cases.
//
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := DeserialiseQuery(tt.args.query); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("DeserialiseQuery() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
