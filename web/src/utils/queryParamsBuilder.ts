import _, { assign, flow, mapValues, omit, partial, pickBy } from "lodash";
import type { BaseQueryCollectionParams } from "types/api";

type EndpointDefinition = {
  endpoint: string;
  baseParam: string;
  suffixEndpoint?: string;
};

const baseQueryDefaultPaginationParams: BaseQueryCollectionParams = {
  limit: 100,
  offset: 0,
};

const parseEndpointDefinition = <T extends { [s: string]: unknown }>(
  { endpoint, baseParam, suffixEndpoint }: EndpointDefinition,
  params: T
): string => {
 if (!suffixEndpoint) return `${endpoint}/${params[baseParam]}`
 return  `${endpoint}/${params[baseParam]}/${suffixEndpoint}`
};

const removeBaseParam = <T extends { [s: string]: unknown }>(endpoint: EndpointDefinition | string, params: T) =>
  typeof endpoint === "object" ? omit(params, endpoint.baseParam) : params;

const removeUndefinedParams = <T extends object>(params: T) => pickBy(params, (value) => value !== undefined);

const mergeQueryWithDefaultPaginationParams = <T>(customParameters: T, isPaginable: boolean): T =>
  isPaginable ? assign(baseQueryDefaultPaginationParams, customParameters) : customParameters;

const stringifyNestedParams = <T extends object>(params: T) =>
  mapValues(params, (value) => (typeof value === "object" ? JSON.stringify(value) : value));

const renderParams = <T>(endpoint: EndpointDefinition | string, params: T, isPaginable: boolean) =>
  flow(
    partial(removeBaseParam, endpoint),
    removeUndefinedParams,
    partial(mergeQueryWithDefaultPaginationParams, _, isPaginable),
    stringifyNestedParams
  )(params);

export const queryParamsBuilder = <T extends { [s: string]: unknown }>(
  endpoint: EndpointDefinition | string,
  params: T,
  isPaginable = false
): string => {
  const parsedEndpoint = typeof endpoint === "object" ? parseEndpointDefinition(endpoint, params) : endpoint;
  const formattedParams = renderParams(endpoint, params, isPaginable);
  const searchParameters = new URLSearchParams(formattedParams);

  const url = `${parsedEndpoint}?${searchParameters.toString()}`;

  return url;
};
