import cellStyles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import { workflowListItem } from "./workflowTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import WorkflowCell from "components/table/cells/channelCell/WorkflowCell";

export default function workflowCellRenderer(
  row: workflowListItem,
  rowIndex: number,
  column: ColumnMetaData<workflowListItem>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: workflowListItem
): JSX.Element {
  if (column.key === "workflowName") {
    return (
      <WorkflowCell
        name={row.workflowName}
        workflowId={row.workflowId}
        workflowVersionId={row.latestWorkflowVersionId}
        className={cellStyles.locked}
        key={"workflow-" + row.workflowId + "-" + rowIndex}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
