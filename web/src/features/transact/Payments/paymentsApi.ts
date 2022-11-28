import { torqApi } from "apiSlice";
import { GetPaymentsQueryParams } from "../../../types/api";
import { queryParamsBuilder } from "../../../utils/queryParamsBuilder";
import { PaymentsResponse } from "./types";

// Define a service using a base URL and expected endpoints
export const api = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getPayments: builder.query<PaymentsResponse, GetPaymentsQueryParams>({
      query: (params) => "payments" + queryParamsBuilder(params, true),
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetPaymentsQuery } = api;
