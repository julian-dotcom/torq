// import { uuid } from "uuidv4";
import { Tag } from "./tagsTypes";
import { ViewResponse } from "features/viewManagement/types";
import { ColumnMetaData } from "features/table/types";

export const TagsColumns: Array<ColumnMetaData<Tag>> = [
  {
    key: "name",
    heading: "Tag",
    type: "TextCell",
    valueType: "string",
  },
];

export const DefaultTagsView: ViewResponse<Tag> = {
  page: "invoices",
  dirty: true,
  view: {
    title: "Draft View",
    columns: TagsColumns,
  },
};
