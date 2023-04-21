import { torqApi } from "apiSlice";
import { NewAddressRequest } from "features/transact/newAddress/newAddressTypes";

// Define a service using a base URL and expected endpoints
export const newAddressApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    newAddress: builder.mutation<string, NewAddressRequest>({
      query: (body) => ({
        url: "/lightning/new-address",
        method: "POST",
        body,
      }),
    }),
  }),
});

export const { useNewAddressMutation } = newAddressApi;
