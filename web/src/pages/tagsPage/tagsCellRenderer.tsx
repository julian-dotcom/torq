import { ExpandedTag } from "./tagsTypes";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TagCell from "components/table/cells/tag/TagCell";
import cellStyles from "components/table/cells/cell.module.scss";
import { TagColor } from "components/tags/Tag";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import { MoleculeRegular as NodeIcon, ArrowRoutingRegular as ChannelIcon } from "@fluentui/react-icons";

import { useDeleteTagMutation } from "./tagsApi";
import useTranslations from "services/i18n/useTranslations";
import { useLocation } from "react-router-dom";
import NumericDoubleCell from "components/table/cells/numeric/NumericDoubleCell";

const color = new Map<string, TagColor>([
  ["error", TagColor.error],
  ["success", TagColor.success],
  ["warning", TagColor.warning],
  ["accent1", TagColor.accent1],
  ["accent2", TagColor.accent2],
  ["accent3", TagColor.accent3],
  ["primary", TagColor.primary],
]);

const categoryColor = new Map<string, TagColor>([
  ["sink", TagColor.error],
  ["balanced", TagColor.success],
  ["source", TagColor.warning],
]);

export default function tagsCellRenderer(
  row: ExpandedTag,
  rowIndex: number,
  column: ColumnMetaData<ExpandedTag>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: ExpandedTag
): JSX.Element {
  const location = useLocation();
  const [deleteTagMutation] = useDeleteTagMutation();
  const { t } = useTranslations();

  if (column.key === "name") {
    return (
      <TagCell
        editLink={true}
        label={row["name"] as string}
        locked={row.tagId ? row.tagId <= 0 : false}
        colorVariant={color.get(row["style"] as string)}
        key={"tag-name" + rowIndex + columnIndex}
        cellWrapperClassName={column.locked ? cellStyles.locked : ""}
        tagId={row.tagId ? row.tagId : 0}
      />
    );
  }

  if (column.key === "delete") {
    return (
      <CellWrapper key={"delete-button-" + row.tagId}>
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

  if (column.key === "channels") {
    return (
      <NumericDoubleCell
        key={"channels-" + rowIndex + columnIndex}
        topValue={row.channels?.length || 0}
        bottomValue={row.nodes?.length || 0}
        topIcon={<ChannelIcon />}
        bottomIcon={<NodeIcon />}
      />
    );
  }

  if (column.key === "edit") {
    return (
      <CellWrapper key={"edit-button-" + row.tagId}>
        <LinkButton
          state={{ background: location }}
          to={`/update-tag/${row.tagId}`}
          buttonSize={SizeVariant.small}
          buttonColor={ColorVariant.primary}
        >
          {t.edit}
        </LinkButton>
      </CellWrapper>
    );
  }

  if (column.key === "categoryId" && row.categoryId) {
    return (
      <TagCell
        key={"category-name-" + row.tagId}
        label={t.tagsModal.tagCategories[row.categoryName]}
        colorVariant={categoryColor.get(row.categoryName as string)}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow, maxRow);
}
