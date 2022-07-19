package payments

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"strings"
)

// TODO: Add fuzzy search
// https://www.crunchydata.com/blog/fuzzy-name-matching-in-postgresql

type FilterClauses struct {
	And    []FilterClauses `json:"$and"`
	Or     []FilterClauses `json:"$or"`
	Filter Filter          `json:"$filter"`
}

type Filter struct {
	FuncName  string `json:"funcName"`
	Category  string `json:"category"`
	Key       string `json:"key"`
	Parameter string `json:"parameter"`
}

func ParseFilter(f Filter) (r sq.Sqlizer) {

	switch f.FuncName {
	case "eq":
		list := strings.Split(f.Parameter, ",")
		if len(list) > 1 {
			return sq.Eq{f.Key: list}
		}
		return sq.Eq{f.Key: f.Parameter}
	case "ne":
		list := strings.Split(f.Parameter, ",")
		if len(list) > 1 {
			return sq.NotEq{f.Key: list}
		}
		return sq.NotEq{f.Key: f.Parameter}
	case "gt":
		return sq.Gt{f.Key: f.Parameter}
	case "gte":
		return sq.GtOrEq{f.Key: f.Parameter}
	case "lt":
		return sq.Lt{f.Key: f.Parameter}
	case "lte":
		return sq.LtOrEq{f.Key: f.Parameter}
	}

	return r
}

func ParseFiltersParams(f FilterClauses) (d sq.Sqlizer) {

	fmt.Println("Runs!")
	if len(f.And) != 0 {
		a := sq.And{}
		for _, v := range f.And {
			a = append(a, ParseFiltersParams(v))
		}
		return a
	}

	if len(f.Or) != 0 {
		a := sq.Or{}
		for _, v := range f.Or {
			a = append(a, ParseFiltersParams(v))
		}
		return a
	}

	return ParseFilter(f.Filter)
}
