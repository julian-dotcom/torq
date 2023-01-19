import { torqApi } from "apiSlice";
import {
  SignMessageRequest,
  SignMessageResponse,
  VerifyMessageRequest,
  VerifyMessageResponse,
} from "./messageVerificationTypes";

export const messageVerificationApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    signMessage: builder.mutation<SignMessageResponse, SignMessageRequest>({
      query: (body) => ({
        url: `messages/sign`,
        method: "POST",
        body: body,
      }),
    }),
    verifyMessage: builder.mutation<VerifyMessageResponse, VerifyMessageRequest>({
      query: (body) => ({
        url: `messages/verify`,
        method: "POST",
        body: body,
      }),
    }),
  }),
});

export const { useSignMessageMutation, useVerifyMessageMutation } = messageVerificationApi;
