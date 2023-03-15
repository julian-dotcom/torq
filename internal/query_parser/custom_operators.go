package query_parser

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
)

// Overlap is used to check if two arrays overlap (SQL operator &&). But can also be used as the ANY
// operator in SQL to check if a string/number is in an array.
func Overlap(param interface{}, key string, notEq bool) (r sq.Sqlizer, err error) {

	var equalOpr = "="

	if notEq {
		equalOpr = "!="
	}

	switch param := param.(type) {
	case nil:
		return sq.Eq{key: nil}, nil
	case []float64:
		sa := sq.Or{}
		for _, value := range param {
			// Here I'm using ARRAY(select (unnest(ARRAY[%s])) to make sure that we have an array in order to use the ANY operator.
			// ARRAY(select (unnest(ARRAY['2'])) means ARRAY['2'].
			// ARRAY(select (unnest(ARRAY[ARRAY['2']])) also means ARRAY['2'].
			// TODO: revisit this custom operator and see if we can solve this in a better way.
			sa = append(sa, sq.Expr(fmt.Sprintf("? %s ANY(ARRAY(select (unnest(ARRAY[%s]))))", equalOpr, key), value))
		}
		return sa, nil
	case []string:
		sa := sq.Or{}
		for _, value := range param {
			sa = append(sa, sq.Expr(fmt.Sprintf("? %s ANY(ARRAY(select (unnest(ARRAY[%s]))))", equalOpr, key), value))
		}
		return sa, nil
	case float64:
		return sq.Expr(fmt.Sprintf("? %s ANY(ARRAY(select (unnest(ARRAY[%s]))))", equalOpr, key), param), nil
	case string:
		return sq.Expr(fmt.Sprintf("? %s ANY(ARRAY(select (unnest(ARRAY[%s]))))", equalOpr, key), param), nil
	default:
		return r, fmt.Errorf("unsupported parameter type: %T. Any only supports string, "+
			"float or an array of string or float", param)
	}
}
