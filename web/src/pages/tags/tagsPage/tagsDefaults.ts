import { ExpandedTag } from "../tagsTypes";
import { ViewResponse } from "features/viewManagement/types";
import { ColumnMetaData } from "features/table/types";

export const TagsColumns: Array<ColumnMetaData<ExpandedTag>> = [
  {
    key: "name",
    heading: "Tag",
    type: "TextCell",
    valueType: "string",
  },
  {
    key: "categoryId",
    heading: "Category",
    type: "TextCell",
    valueType: "string",
  },
  { key: "channels", heading: "Applied to", type: "NumericDoubleCell", valueType: "string" },
  {
    key: "edit",
    heading: "Edit",
    type: "EditCell",
    valueType: "string",
  },
  {
    key: "delete",
    heading: "Delete",
    type: "DeleteCell",
    valueType: "string",
  },
];

export const DefaultTagsView: ViewResponse<ExpandedTag> = {
  page: "invoices",
  dirty: true,
  view: {
    title: "Draft View",
    columns: TagsColumns,
  },
};
