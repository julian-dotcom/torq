package workflows

import (
	"reflect"
	"testing"
)

func TestDeserialiseQuery(t *testing.T) {
	t.Run("returns empty Clause for nil query", func(t *testing.T) {
		result := DeserialiseQuery(nil)
		expected := Clause{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got %v, expected %v", result, expected)
		}
	})

	t.Run("returns Clause with empty $and filter", func(t *testing.T) {
		query := FilterClauses{
			And: []FilterClauses{},
		}
		result := DeserialiseQuery(query)
		expected := Clause{
			Prefix: "$and",
			Filter: FilterInterface{},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got %v, expected %v", result, expected)
		}
	})

	t.Run("returns Clause with Filter for FilterClauses query", func(t *testing.T) {
		query := FilterClauses{
			Filter: Filter{
				Parameter: "param",
				FuncName:  "func",
				Key:       "key",
				Category:  "date",
			},
		}
		result := DeserialiseQuery(query)
		expected := Clause{
			Prefix: "$filter",
			Filter: FilterInterface{
				Parameter: "param",
				FuncName:  "func",
				Key:       "key",
				Category:  FilterCategoryTypeDate,
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got \n%v, \nexpected \n%v", result, expected)
		}
	})

	t.Run("returns Clause with ChildClauses for And query", func(t *testing.T) {
		query := FilterClauses{
			And: []FilterClauses{
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param1",
				}},
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param2",
				}},
			},
		}
		result := DeserialiseQuery(query)
		expected := Clause{
			Prefix: "$and",
			ChildClauses: []Clause{
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param1", FuncName: "func"},
					Result:       false,
				},
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param2", FuncName: "func"},
					Result:       false,
				},
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got \n%v, \nexpected \n%v", result, expected)
		}
	})

	t.Run("returns Clause with ChildClauses for Or query", func(t *testing.T) {
		query := FilterClauses{
			Or: []FilterClauses{
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param1",
				}},
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param2",
				}},
			},
		}
		result := DeserialiseQuery(query)
		expected := Clause{
			Prefix: "$or",
			ChildClauses: []Clause{
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param1", FuncName: "func"},
					Result:       false,
				},
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param2", FuncName: "func"},
					Result:       false,
				},
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got \n%v, \nexpected \n%v", result, expected)
		}
	})

	t.Run("returns nested Clause with ChildClauses", func(t *testing.T) {
		query := FilterClauses{
			And: []FilterClauses{
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param1",
				}},
				{Filter: Filter{
					FuncName:  "func",
					Parameter: "param2",
				}},
				{
					Or: []FilterClauses{
						{Filter: Filter{
							FuncName:  "func",
							Parameter: "param1",
						}},
						{Filter: Filter{
							FuncName:  "func",
							Parameter: "param2",
						}},
					},
				},
			},
		}
		result := DeserialiseQuery(query)
		expected := Clause{
			Prefix: "$and",
			ChildClauses: []Clause{
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param1", FuncName: "func"},
					Result:       false,
				},
				{
					Prefix:       "$filter",
					ChildClauses: nil,
					Filter:       FilterInterface{Parameter: "param2", FuncName: "func"},
					Result:       false,
				},
				{
					Prefix: "$or",
					ChildClauses: []Clause{
						{
							Prefix:       "$filter",
							ChildClauses: nil,
							Filter:       FilterInterface{Parameter: "param1", FuncName: "func"},
							Result:       false,
						},
						{
							Prefix:       "$filter",
							ChildClauses: nil,
							Filter:       FilterInterface{Parameter: "param2", FuncName: "func"},
							Result:       false,
						},
					},
				},
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("got \n%v, \nexpected \n%v", result, expected)
		}
	})

	t.Run("panics for unexpected query format", func(t *testing.T) {
		query := struct{}{}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic, but did not panic")
			}
		}()
		DeserialiseQuery(query)
	})
}
