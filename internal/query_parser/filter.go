package query_parser

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

//var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
//var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
//
//func ToSnakeCase(str string) string {
//	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
//	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
//	return strings.ToLower(snake)
//}

// Examples of json input
//
// Example 1:
// {"$filter":{"funcName":"eq","key":"status","parameter":"SUCCEEDED"}}
//
// Example 2:
// {"$and":[
//  {"$filter":{"funcName":"eq","key":"status","parameter":"SUCCEEDED"}},
//  {"$filter":{"funcName":"gte","key":"amount_msat","parameter":2000}}
// ]}
//
// Example 3:
// {"$or":[
//  {"$filter":{"funcName":"eq","key":"status","parameter":"SUCCEEDED"}},
//  {"$filter":{"funcName":"gte","key":"amount_msat","parameter":2000}}
// ]}
//
// Example 4:
// {"$or":[
//   {"$and":[
//    {"$filter":{"funcName":"eq","key":"status","parameter":"SUCCEEDED"}},
//    {"$filter":{"funcName":"gte","key":"amount_msat","parameter":2000}}
//   ]},
//   {"$and":[
//    {"$filter":{"funcName":"eq","key":"status","parameter":"FAILED"}},
//    {"$filter":{"funcName":"lt","key":"amount_msat","parameter":1000}}
//   ]}
// ]}

func ParseFilterParam(params string, allowedColumns []string) (f sq.Sqlizer, err error) {

	filters := FilterClauses{}
	err = json.Unmarshal([]byte(params), &filters)
	if err != nil {
		return f, err
	}

	qp := QueryParser{
		AllowedColumns: allowedColumns,
	}
	f, err = qp.ParseFilterClauses(filters)
	if err != nil {
		return f, err
	}

	return f, nil
}

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
}

func (qp *QueryParser) ParseFilter(f Filter) (r sq.Sqlizer, err error) {

	//key, err := GetDBKeyName(f.Key)
	//if err != nil {
	//	return r, err
	//}
	key := strcase.ToSnake(f.Key)
	if !qp.IsAllowed(key) {
		return r,
			fmt.Errorf("filtering by %s is not allwed. Try one of: %v",
				key,
				strings.Join(qp.AllowedColumns, ", "),
			)
	}

	param := f.Parameter

	switch param.(type) {
	case string:
		break
	case float64:
		break
	case bool:
		break
	case []interface{}:
		l := f.Parameter.([]interface{})
		var paramList []string
		for _, v := range l {
			paramList = append(paramList, fmt.Sprintf("%v", v))
		}
		param = paramList
	default:
		return r, fmt.Errorf("unsupported parameter type: %T", f.Parameter)
	}

	switch f.FuncName {
	case "eq":
		return sq.Eq{key: param}, nil
	case "neq":
		return sq.NotEq{key: param}, nil
	case "gt":
		return sq.Gt{key: param}, nil
	case "gte":
		return sq.GtOrEq{key: param}, nil
	case "lt":
		return sq.Lt{key: param}, nil
	case "lte":
		return sq.LtOrEq{key: param}, nil
	case "like":
		return sq.ILike{key: "%" + fmt.Sprintf("%v", param) + "%"}, nil
	case "notLike":
		return sq.NotILike{key: "%" + fmt.Sprintf("%v", param) + "%"}, nil
	case "any":
		return Overlap(param, key, false)
	case "notAny":
		return Overlap(param, key, true)
	default:
		return r, fmt.Errorf("%s is not a valid filter function", f.FuncName)
	}
}

func (qp *QueryParser) ParseFilterClauses(f FilterClauses) (d sq.Sqlizer, err error) {

	if len(f.And) != 0 {
		a := sq.And{}
		for _, v := range f.And {
			r, err := qp.ParseFilterClauses(v)
			if err != nil {
				return d, err
			}
			a = append(a, r)
		}
		return a, nil
	}

	if len(f.Or) != 0 {
		a := sq.Or{}
		for _, v := range f.Or {
			r, err := qp.ParseFilterClauses(v)
			if err != nil {
				return d, err
			}
			a = append(a, r)
		}
		return a, nil
	}

	return qp.ParseFilter(f.Filter)
}
