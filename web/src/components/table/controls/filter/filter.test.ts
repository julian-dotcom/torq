import { AndClause, OrClause, FilterClause, processQuery, deserialiseQueryJSON, filterFuncNameType } from "./filter"

const data = { capacity: 99 }

const failingFilter = {
  combiner: "and" as "and" | "or",
  funcName: "gte" as filterFuncNameType,
  category: "number" as "number" | "string",
  key: "capacity",
  parameter: 100
}

const passingFilter = {
  combiner: "and" as "and" | "or",
  funcName: "gte" as filterFuncNameType,
  category: "number" as "number" | "string",
  key: "capacity",
  parameter: 99
}

test('simplest query returns false', () => {
  const simplestClauseFalse = new FilterClause(failingFilter)
  const result = processQuery(simplestClauseFalse, data)
  expect(result).toBe(false);
})

test('simplest AND query returns false', () => {
  const simpleAndClauseReturnFalse = new AndClause()
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(failingFilter))
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(passingFilter))
  const result = processQuery(simpleAndClauseReturnFalse, data)
  expect(result).toBe(false);
})

test('simplest AND query returns true', () => {
  const simpleAndClauseReturnTrue = new AndClause()
  simpleAndClauseReturnTrue.addChildClause(new FilterClause(passingFilter))
  simpleAndClauseReturnTrue.addChildClause(new FilterClause(passingFilter))
  const result = processQuery(simpleAndClauseReturnTrue, data)
  expect(result).toBe(true);
})

test('simplest OR query returns true', () => {
  const clauseOrReturnTrue = new OrClause()
  clauseOrReturnTrue.addChildClause(new FilterClause(passingFilter))
  clauseOrReturnTrue.addChildClause(new FilterClause(failingFilter))
  const result = processQuery(clauseOrReturnTrue, data)
  expect(result).toBe(true);
})

test('simplest OR query returns false', () => {
  const clauseOrReturnFalse = new OrClause()
  clauseOrReturnFalse.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnFalse.addChildClause(new FilterClause(failingFilter))
  const result = processQuery(clauseOrReturnFalse, data);
  expect(result).toBe(false);
})

test('complex query returns true', () => {
  const simpleAndClauseReturnFalse = new AndClause()
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(failingFilter))
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(passingFilter))

  const simpleAndClauseReturnTrue = new AndClause()
  simpleAndClauseReturnTrue.addChildClause(new FilterClause(passingFilter))
  simpleAndClauseReturnTrue.addChildClause(new FilterClause(passingFilter))

  const multiClauseReturnTrue = new OrClause()
  multiClauseReturnTrue.addChildClause(simpleAndClauseReturnTrue)
  multiClauseReturnTrue.addChildClause(simpleAndClauseReturnFalse)

  const result = processQuery(multiClauseReturnTrue, data);
  expect(result).toBe(true);
})

test('complex query returns false', () => {
  const simpleAndClauseReturnFalse = new AndClause()
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(failingFilter))
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(passingFilter))

  const clauseOrReturnFalse = new OrClause()
  clauseOrReturnFalse.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnFalse.addChildClause(new FilterClause(failingFilter))

  const multiClauseReturnFalse = new OrClause()
  multiClauseReturnFalse.addChildClause(clauseOrReturnFalse)
  multiClauseReturnFalse.addChildClause(simpleAndClauseReturnFalse)

  const result = processQuery(multiClauseReturnFalse, data);
  expect(result).toBe(false);
})

// test('simplest query serialises itself', () => {
//   const expected = {
//     "$filter": false
//   }
//   const simplestClauseFalse = new FilterClause(false)
//   const result = JSON.parse(JSON.stringify(simplestClauseFalse))
//   expect(result).toEqual(expected);
// })

// test('simple AND query serialises itself', () => {
//   const expected = {
//     $and: [
//       { $filter: false },
//       { $filter: true },
//     ]
//   }
//   const simpleAndClauseReturnFalse = new AndClause()
//   simpleAndClauseReturnFalse.addChildClause(new FilterClause(false))
//   simpleAndClauseReturnFalse.addChildClause(new FilterClause(true))

//   const result = JSON.parse(JSON.stringify(simpleAndClauseReturnFalse))
//   expect(result).toEqual(expected);
// })


// test('complex query serialises itself', () => {
//   const expected = {
//     $and: [
//       { $filter: false },
//       {
//         $or: [
//           { $filter: false },
//           { $filter: false },
//           { $filter: true }]
//       }
//     ]
//   }

//   const clauseOrReturnTrue = new OrClause()
//   clauseOrReturnTrue.addChildClause(new FilterClause(false))
//   clauseOrReturnTrue.addChildClause(new FilterClause(false))
//   clauseOrReturnTrue.addChildClause(new FilterClause(true))

//   const complexClause = new AndClause()
//   complexClause.addChildClause(new FilterClause(false))
//   complexClause.addChildClause(clauseOrReturnTrue)

//   const result = JSON.parse(JSON.stringify(complexClause))
//   expect(result).toEqual(expected);
// })

// test('deserialise into simple query', () => {
//   const simpleQuery = {
//     "$filter": false
//   }
//   const simpleQueryJSON = JSON.stringify(simpleQuery)
//   const result = deserialiseQueryJSON(simpleQueryJSON)
//   const expected = new FilterClause(false)
//   expect(result).toEqual(expected);
// })

// test('deserialise a complex query', () => {
//   const complexQuery = {
//     $and: [
//       { $filter: false },
//       {
//         $or: [
//           { $filter: false },
//           { $filter: false },
//           { $filter: true }]
//       }
//     ]
//   }

//   const clauseOrReturnTrue = new OrClause()
//   clauseOrReturnTrue.addChildClause(new FilterClause(false))
//   clauseOrReturnTrue.addChildClause(new FilterClause(false))
//   clauseOrReturnTrue.addChildClause(new FilterClause(true))

//   const complexClause = new AndClause()
//   complexClause.addChildClause(new FilterClause(false))
//   complexClause.addChildClause(clauseOrReturnTrue)

//   const complexQueryJSON = JSON.stringify(complexQuery)
//   const result = deserialiseQueryJSON(complexQueryJSON)
//   expect(result).toEqual(complexClause);
// })
