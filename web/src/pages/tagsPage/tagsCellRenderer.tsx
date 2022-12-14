import { Tag } from "./tagsTypes";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TagCell from "components/table/cells/tag/TagCell";
import cellStyles from "components/table/cells/cell.module.scss";
import { TagColor } from "../../components/tags/Tag";

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
      ["drain", TagColor.error],
      ["source", TagColor.success],
      ["routing", TagColor.warning],
    ]);

    return (
      <TagCell
        label={row["name"] as string}
        locked={row["tagId"] <= 0}
        colorVariant={color.get(row["name"] as string)}
        key={"tag-name" + rowIndex + columnIndex}
        cellWrapperClassName={column.locked ? cellStyles.locked : ""}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow, maxRow);
}
