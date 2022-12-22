import { useId, useState } from "react";
import {
  PuzzlePiece20Regular as NodesIcon,
  Play20Regular as DeployIcon,
  Save20Regular as SaveDraftIcon,
} from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import ChannelPolicyNode from "components/workflow/nodes/ChannelPolicy";
import { Link, useParams } from "react-router-dom";
import { useGetWorkflowQuery } from "./workflowApi";
import Sidebar from "features/sidebar/Sidebar";
import classNames from "classnames";
import Button, { buttonColor } from "components/buttons/Button";
import { WORKFLOWS, MANAGE } from "constants/routes";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";

type WorkflowPageProps = {
  title?: string;
  workflowId?: string;
  breadcrumbs?: Array<any>;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const { workflowId, version } = useParams();

  const { data } = useGetWorkflowQuery(
    {
      workflowId: parseInt(workflowId || ""),
      version: parseInt(version || ""),
    },
    { skip: !workflowId || !version }
  );

  const bradcrumbs = props.breadcrumbs || [
    <Link to={`/${MANAGE}/${WORKFLOWS}`} key={"workflowsLink"}>
      {t.workflows}
    </Link>,
    data?.workflow?.name,
    data?.version?.name,
  ];
  const id1 = useId();
  const id2 = useId();
  const id3 = useId();

  const nodes = new Map([
    ["node1", <ChannelPolicyNode id={id1} key={"node1"} nodeName={"firstNode"} />],
    ["node2", <ChannelPolicyNode id={id2} key={"node2"} nodeName={"secondNode"} />],
    ["node3", <ChannelPolicyNode id={id3} key={"node3"} nodeName={"thirdNode"} />],
  ]);

  const [sidebarExpanded, setSidebarExpanded] = useState(true);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
  };

  const workflowControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={buttonColor.green}
            text={t.deploy}
            className={"collapse-tablet"}
            icon={<DeployIcon />}
            onClick={() => {
              console.log("Not implemented yet");
            }}
          />
          <Button
            buttonColor={buttonColor.green}
            text={t.saveDraft}
            className={"collapse-tablet"}
            icon={<SaveDraftIcon />}
            onClick={() => {
              console.log("Not implemented yet");
            }}
          />
        </TableControlsTabsGroup>
        <Button
          buttonColor={buttonColor.subtle}
          text={t.nodes}
          className={"collapse-tablet"}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            setSidebarExpanded(!sidebarExpanded);
          }}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={(data && data.workflow.name) || ""} />
      {workflowControls}
      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>
            <WorkflowCanvas>{Array.from(nodes.values())}</WorkflowCanvas>
          </div>
        </div>
      </div>
      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: sidebarExpanded })}>
        <Sidebar title={t.nodes} closeSidebarHandler={closeSidebarHandler}>
          <div>{"nodes here"}</div>
        </Sidebar>
      </div>
    </div>
  );
}

export default WorkflowPage;
