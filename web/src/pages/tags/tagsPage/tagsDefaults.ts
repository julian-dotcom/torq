import { ExpandedTag } from "pages/tags/tagsTypes";
import { ViewResponse } from "features/viewManagement/types";
import { AllTagsColumns } from "pages/tags/tagsPage/tagsColumns";

export const DefaultTagsView: ViewResponse<ExpandedTag> = {
  page: "invoices",
  dirty: true,
  view: {
    title: "Draft View",
    columns: AllTagsColumns,
  },
};
