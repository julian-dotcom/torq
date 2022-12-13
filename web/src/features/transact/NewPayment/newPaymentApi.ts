import { torqApi } from "apiSlice";
import { DecodedInvoice, GetDecodedInvoiceQueryParams, SendOnChainRequest, SendOnChainResponse } from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";

// Define a service using a base URL and expected endpoints
export const NewPaymentApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getDecodedInvoice: builder.query<DecodedInvoice, GetDecodedInvoiceQueryParams>({
      query: (params) => "invoices/decode" + queryParamsBuilder(params),
    }),
    sendOnChain: builder.mutation<SendOnChainResponse, SendOnChainRequest>({
      query: (data: SendOnChainRequest) => ({
        url: "on-chain-tx/sendcoins",
        method: "POST",
        body: data,
      }),
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetDecodedInvoiceQuery, useSendOnChainMutation } = NewPaymentApi;
