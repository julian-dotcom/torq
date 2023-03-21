import { torqApi } from "apiSlice";
import { GetOnChainTransactionsQueryParams } from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { OnChainResponse } from "./types";

// Define a service using a base URL and expected endpoints
export const onChainApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getOnChainTx: builder.query<OnChainResponse, GetOnChainTransactionsQueryParams>({
      query: (params) => "on-chain-tx" + queryParamsBuilder(params, true),
      providesTags: ["onChainTx"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetOnChainTxQuery } = onChainApi;
