import { Tag } from "./tagsTypes";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TagCell from "components/table/cells/tag/TagCell";
import cellStyles from "components/table/cells/cell.module.scss";
import { TagColor } from "components/tags/Tag";
// import { Link } from "react-router-dom";

export default function tagsCellRenderer(
  row: Tag,
  rowIndex: number,
  column: ColumnMetaData<Tag>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Tag
): JSX.Element {
  if (column.key === "name") {
    const color = new Map<string, TagColor>([
      ["error", TagColor.error],
      ["success", TagColor.success],
      ["warning", TagColor.warning],
      ["accent1", TagColor.accent1],
      ["accent2", TagColor.accent2],
      ["accent3", TagColor.accent3],
      ["primary", TagColor.primary],
    ]);

    return (
      <TagCell
        label={row["name"] as string}
        locked={row.tagId ? row.tagId <= 0 : false}
        colorVariant={color.get(row["style"] as string)}
        key={"tag-name" + rowIndex + columnIndex}
        cellWrapperClassName={column.locked ? cellStyles.locked : ""}
        tagId={row.tagId ? row.tagId : 0}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow, maxRow);
}
