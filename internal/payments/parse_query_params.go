package payments

import (
	sq "github.com/Masterminds/squirrel"
	"strings"
)

func ParseFilter(f Filter) (r sq.Sqlizer) {

	switch f.FuncName {
	case "eq":
		list := strings.Split(f.Parameter, ",")
		if len(list) > 1 {
			return sq.Eq{f.Key: f.Parameter}
		}
	case "ne":
		list := strings.Split(f.Parameter, ",")
		if len(list) > 1 {
			return sq.NotEq{f.Key: f.Parameter}
		}
	case "gt":
		return sq.Gt{f.Key: f.Parameter}
	case "gte":
		return sq.GtOrEq{f.Key: f.Parameter}
	case "lt":
		return sq.Lt{f.Key: f.Parameter}
	case "lte":
		return sq.LtOrEq{f.Key: f.Parameter}

		//	"ne":    sq.NotEq,
		//"gt":    sq.Gt,
		//"gte":   sq.GtOrEq,
		//"lt":    sq.Lt,
		//"lte":   sq.LtOrEq,
		//"like":  sq.Like,
		//"nlike": sq.NotLike,
	}

	return r
}

func ParseFiltersParams(f FilterClauses) (d []sq.Sqlizer) {

	for _, v := range f.And {
		a := ParseFiltersParams(v)
		return sq.And(a)
	}

	for _, v := range f.Or {
		a := ParseFiltersParams(v)
		return sq.Or(a)
	}

	//for combiner, filter := range f {
	//
	//	switch combiner {
	//	case "$and":
	//
	//		dr, err := ParseFiltersParams(filter)
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		return sq.And(dr), nil
	//
	//	case "or":
	//		dr, err := ParseFiltersParams(filter)
	//		if err != nil {
	//			return nil, err
	//		}
	//		return sq.Or(dr), nil
	//
	//	case "$filter":
	//		dr, err := ParseFilter(filter)
	//		if err != nil {
	//			return nil, err
	//		}
	//		d = append(d, dr)
	//	}

	return d
}

//func parseClause(query: , data: any) {
//  switch (clause.prefix) {
//    case "$filter": {
//      const filterClause = clause as FilterClause
//      const filterFunc = FilterFunctions.get(filterClause.filter.category)?.get(filterClause.filter.funcName)
//      if (!filterFunc) {
//        throw new Error("Filter function is not yet defined")
//      }
//      clause.result = filterFunc(data, filterClause.filter.key, filterClause.filter.parameter)
//      break;
//    }
//    case "$and": {
//      for (const childClause of (clause as AndClause).childClauses) {
//        // recursive call processing each child clause
//        parseClause(childClause, data)
//        // if any of the sibling filters are false then the AND fails, no need to process the rest
//        if ((childClause as ClauseWithResult).result === false) {
//          clause.result = false
//          break typeSwitch
//        }
//      }
//      // check that every filter is true so satisfy the AND
//      if ((clause as AndClause).childClauses.every(sc => (sc as ClauseWithResult).result === true)) {
//        clause.result = true
//      }
//      break;
//    }
//    case "$or": {
//      for (const childClause of (clause as OrClause).childClauses) {
//        // recursive call processing each child clause
//        parseClause(childClause, data)
//        // if any of the sibling filters are true then the OR succeeds, no need to process the rest
//        if ((childClause as ClauseWithResult).result === true) {
//          clause.result = true
//          break typeSwitch
//        }
//      }
//      // if we made it here all of the previous filters must have returned false so whole OR fails
//      clause.result = false
//      break;
//    }
//  }
//}
