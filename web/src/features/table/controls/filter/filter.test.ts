import { AndClause, OrClause, FilterClause, processQuery, deserialiseQueryFromString } from "./filter"

const data = { capacity: 99 }

const failingFilter = {
  combiner: "and" as "and" | "or",
  funcName: "gte",
  category: "number" as "number" | "string",
  key: "capacity",
  parameter: 100
}

const passingFilter = {
  combiner: "and" as "and" | "or",
  funcName: "gte",
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

test('simplest query serialises itself', () => {
  const expected = {
    "$filter": failingFilter
  }
  const simplestClauseFalse = new FilterClause(failingFilter)
  const result = JSON.parse(JSON.stringify(simplestClauseFalse))
  expect(result).toEqual(expected);
})

test('simple AND query serialises itself', () => {
  const expected = {
    $and: [
      { $filter: failingFilter },
      { $filter: passingFilter },
    ]
  }
  const simpleAndClauseReturnFalse = new AndClause()
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(failingFilter))
  simpleAndClauseReturnFalse.addChildClause(new FilterClause(passingFilter))

  const result = JSON.parse(JSON.stringify(simpleAndClauseReturnFalse))
  expect(result).toEqual(expected);
})

test('complex query serialises itself', () => {
  const expected = {
    $and: [
      { $filter: failingFilter },
      {
        $or: [
          { $filter: failingFilter },
          { $filter: failingFilter },
          { $filter: passingFilter }]
      }
    ]
  }

  const clauseOrReturnTrue = new OrClause()
  clauseOrReturnTrue.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnTrue.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnTrue.addChildClause(new FilterClause(passingFilter))

  const complexClause = new AndClause()
  complexClause.addChildClause(new FilterClause(failingFilter))
  complexClause.addChildClause(clauseOrReturnTrue)

  const result = JSON.parse(JSON.stringify(complexClause))
  expect(result).toEqual(expected);
})

test('deserialise into simple query', () => {
  const simpleQuery = {
    "$filter": failingFilter
  }
  const simpleQueryAsString = JSON.stringify(simpleQuery)
  const result = deserialiseQueryFromString(simpleQueryAsString)
  const expected = new FilterClause(failingFilter)
  expect(result).toEqual(expected);
})

test('deserialise a complex query', () => {
  const complexQuery = {
    $and: [
      { $filter: failingFilter },
      {
        $or: [
          { $filter: failingFilter },
          { $filter: failingFilter },
          { $filter: passingFilter }]
      }
    ]
  }

  const clauseOrReturnTrue = new OrClause()
  clauseOrReturnTrue.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnTrue.addChildClause(new FilterClause(failingFilter))
  clauseOrReturnTrue.addChildClause(new FilterClause(passingFilter))

  const complexClause = new AndClause()
  complexClause.addChildClause(new FilterClause(failingFilter))
  complexClause.addChildClause(clauseOrReturnTrue)

  const complexQueryAsString = JSON.stringify(complexQuery)
  const result = deserialiseQueryFromString(complexQueryAsString)
  expect(result).toEqual(complexClause);
})
