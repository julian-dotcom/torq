import { Add16Regular as NewWorkflowIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import TablePageTemplate, {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";

import Table from "features/table/Table";
import { ColumnMetaData } from "features/table/types";
import { workflowListItem } from "./workflowTypes";
import workflowCellRenderer from "./workflowCellRenderer";
import { useGetWorkflowsQuery, useNewWorkflowMutation } from "./workflowApi";
import Button, { buttonColor } from "../../components/buttons/Button";
import { useNavigate } from "react-router";

// type WorkflowsTablePageProps = {};

function WorkflowsTablePage() {
  const { t } = useTranslations();
  const breadcrumbs = [t.manage, t.workflows];
  const navigate = useNavigate();

  const workflowListResponse = useGetWorkflowsQuery();
  const [newWorkflow] = useNewWorkflowMutation();

  function newWorkflowHandler() {
    const response = newWorkflow();
    response
      .then((res) => {
        console.log(res);
        const data = (res as { data: { workflowId: number; version: number } }).data;
        navigate(`/manage/workflows/${data.workflowId}/versions/${data.version}`);
      })
      .catch((err) => {
        console.log(err);
      });
  }

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

  const workflowControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={buttonColor.green}
            text={t.newWorkflow}
            className={"collapse-tablet"}
            icon={<NewWorkflowIcon />}
            onClick={newWorkflowHandler}
          />
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  return (
    <TablePageTemplate title={t.workflows} breadcrumbs={breadcrumbs} tableControls={workflowControls}>
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
