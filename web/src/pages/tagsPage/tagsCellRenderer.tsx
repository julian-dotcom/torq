import { ExpandedTag, Tag } from "./tagsTypes";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TagCell from "components/table/cells/tag/TagCell";
import cellStyles from "components/table/cells/cell.module.scss";
import { TagColor } from "components/tags/Tag";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useDeleteTagMutation } from "./tagsApi";
import useTranslations from "services/i18n/useTranslations";

export default function tagsCellRenderer(
  row: Tag,
  rowIndex: number,
  column: ColumnMetaData<ExpandedTag>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Tag
): JSX.Element {
  const [deleteTagMutation] = useDeleteTagMutation();
  const { t } = useTranslations();

  if (column.key === "delete") {
    return (
      <CellWrapper>
        <Button
          disabled={!(row.tagId !== undefined && row.tagId >= 0)}
          buttonSize={SizeVariant.small}
          onClick={() => {
            deleteTagMutation(row.tagId || 0);
          }}
          buttonColor={ColorVariant.error}
        >
          {t.delete}
        </Button>
      </CellWrapper>
    );
  }
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
