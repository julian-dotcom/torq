package query_parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)
type Order struct {
	Key       string `json:"key"`
	Direction string `json:"direction"`
}

func ParseOrderParams(params string, allowedColumns []string) ([]string, error) {
	var sort []Order
	err := json.Unmarshal([]byte(params), &sort)
	if err != nil {
		return nil, err
	}

	for i, param := range sort {
		sort[i].Key =  strcase.ToSnake(param.Key)
	}

	// Whitelist the columns that are allowed to sorted by.
	qp := QueryParser{
		AllowedColumns: allowedColumns,
	}

	sortString, err := qp.ParseOrderClauses(sort)
	if err != nil {
		return nil, err
	}
	return sortString, nil
}

func (qp *QueryParser) ParseOrder(s Order) (r string, err error) {

	// Prevents SQL injection by only allowing whitelisted column names.
	if !qp.IsAllowed(s.Key) {
		return r,
			fmt.Errorf("sorting by %s is not allwed. Try one of: %v",
				s.Key,
				strings.Join(qp.AllowedColumns, ", "),
			)
	}

	// Prevent SQL injection by only allowing asc and desc as directions.
	if !(s.Direction == "asc" || s.Direction == "desc") {
		return r, fmt.Errorf("%s is not a valid sort direction. Should be either asc or desc", s.Direction)
	}

	return fmt.Sprintf("%s %s", s.Key, s.Direction), nil
}

func (qp *QueryParser) ParseOrderClauses(s []Order) (r []string, err error) {

	for _, sort := range s {
		os, err := qp.ParseOrder(sort)
		if err != nil {
			return r, err
		}
		r = append(r, os)
	}

	return r, nil
}
