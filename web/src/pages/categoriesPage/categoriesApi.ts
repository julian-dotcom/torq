import { torqApi } from "apiSlice";
import { Category } from "./categoriesTypes";

// Define a service using a base URL and expected endpoints
export const onChainApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getCategories: builder.query<Category[], void>({
      query: () => `categories`,
      providesTags: ["categories"],
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const { useGetCategoriesQuery } = onChainApi;
