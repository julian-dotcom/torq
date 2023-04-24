import { torqApi } from "apiSlice";
import { NewInvoiceRequest, NewInvoiceResponse } from "./newInvoiceTypes";

// Define a service using a base URL and expected endpoints
export const invoiceApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    newInvoice: builder.mutation<NewInvoiceResponse, NewInvoiceRequest>({
      query: (data: NewInvoiceRequest) => {
        return {
          url: "lightning/newinvoice",
          method: "POST",
          body: data,
        };
      },
      invalidatesTags: ["invoices"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useNewInvoiceMutation } = invoiceApi;
