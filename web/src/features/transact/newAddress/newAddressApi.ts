import { torqApi } from "apiSlice";
import { NewAddressRequest, NewAddressResponse } from "./newAddressTypes";

// Define a service using a base URL and expected endpoints
export const newAddressApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    newAddress: builder.mutation<NewAddressResponse, NewAddressRequest>({
      query: (body) => ({
        url: "/on-chain-tx/new-address",
        method: "POST",
        body,
      }),
    }),
  }),
});

export const { useNewAddressMutation } = newAddressApi;
