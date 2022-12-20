import useTranslations from "services/i18n/useTranslations";
import TablePageTemplate from "features/templates/tablePageTemplate/TablePageTemplate";
import Table from "features/table/Table";
import { ColumnMetaData } from "features/table/types";
import { workflowListItem } from "./workflowTypes";
import workflowCellRenderer from "./workflowCellRenderer";
import { useGetWorkflowsQuery } from "./workflowApi";

// type WorkflowsTablePageProps = {};

function WorkflowsTablePage() {
  const { t } = useTranslations();
  const breadcrumbs = [t.manage, t.workflows];

  const workflowListResponse = useGetWorkflowsQuery();

  const columns: Array<ColumnMetaData<workflowListItem>> = [
    {
      key: "workflowName",
      heading: "Name",
      valueType: "string",
      type: "TextCell",
    },
    {
      key: "workflowStatus",
      heading: "Active",
      valueType: "boolean",
      type: "BooleanCell",
    },
    {
      key: "activeVersionName",
      heading: "Active Version",
      valueType: "string",
      type: "TextCell",
    },
  ];

  return (
    <TablePageTemplate title={t.workflows} breadcrumbs={breadcrumbs} tableControls={<div />}>
      <Table
        cellRenderer={workflowCellRenderer}
        data={workflowListResponse.data || []}
        activeColumns={columns}
        isLoading={false}
      />
    </TablePageTemplate>
  );
}

export default WorkflowsTablePage;
