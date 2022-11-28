import { torqApi } from "apiSlice";
import { GetInvoicesQueryParams } from "types/api";
import { queryParamsBuilder } from "utils/queryParamsBuilder";
import { InvoicesResponse } from "./invoiceTypes";

// Define a service using a base URL and expected endpoints
export const invoiceApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getInvoices: builder.query<InvoicesResponse, GetInvoicesQueryParams>({
      query: (params) => "invoices" + queryParamsBuilder(params, true),
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetInvoicesQuery } = invoiceApi;
