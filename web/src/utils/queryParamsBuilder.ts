import { assign, mapValues, pickBy } from "lodash";
import type { BaseQuery } from "types/api";

const baseQueryDefaultPaginationParams: Partial<BaseQuery> = {
  limit: 100,
  offset: 0,
};

const mergeQueryWithDefaultPaginationParams = (defaultParameters: Partial<BaseQuery>, customParameters: BaseQuery) =>
  assign(defaultParameters, customParameters);

export const queryParamsBuilder = <T extends BaseQuery>(endpoint: string, params: T, isPaginable = false): string => {
  const mergeDefaultPaginationParams = isPaginable
    ? mergeQueryWithDefaultPaginationParams(baseQueryDefaultPaginationParams, params)
    : params;

  const stringifyNested = mapValues(mergeDefaultPaginationParams, (value) =>
    typeof value === "object" ? JSON.stringify(value) : value
  );

  const removeUndefined = pickBy(stringifyNested, (value) => value !== undefined);

  const searchParameters = new URLSearchParams(removeUndefined as any);

  const url = `${endpoint}?${searchParameters.toString()}`;

  return url;
};
