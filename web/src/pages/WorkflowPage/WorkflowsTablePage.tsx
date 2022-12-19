import useTranslations from "services/i18n/useTranslations";
import TablePageTemplate from "features/templates/tablePageTemplate/TablePageTemplate";
import Table from "features/table/Table";
import { ColumnMetaData } from "features/table/types";
import { workflowListItem } from "./workflowTypes";
import workflowCellRenderer from "./workflowCellRenderer";

// type WorkflowsTablePageProps = {};

function WorkflowsTablePage() {
  const { t } = useTranslations();
  const breadcrumbs = [t.manage, t.workflows];

  const data = [
    {
      id: 1,
      name: "Workflow 1",
    },
    {
      id: 2,
      name: "Workflow 2",
    },
  ];

  const columns: Array<ColumnMetaData<workflowListItem>> = [
    // {
    //   key: "id",
    //   heading: "ID",
    //   valueType: "number",
    //   type: "NumberCell",
    // },
    {
      key: "name",
      heading: "Name",
      valueType: "string",
      type: "TextCell",
    },
  ];

  return (
    <TablePageTemplate title={t.workflows} breadcrumbs={breadcrumbs} tableControls={<div />}>
      <Table cellRenderer={workflowCellRenderer} data={data} activeColumns={columns} isLoading={false} />
    </TablePageTemplate>
  );
}

export default WorkflowsTablePage;
