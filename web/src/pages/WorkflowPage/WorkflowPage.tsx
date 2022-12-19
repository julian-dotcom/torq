import { useId } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import ChannelPolicyNode from "components/workflow/nodes/ChannelPolicy";
import { useParams } from "react-router-dom";

type WorkflowPageProps = {
  title?: string;
  workflowId?: string;
  breadcrumbs?: Array<any>;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const { workflowId } = useParams();

  const bradcrumbs = props.breadcrumbs || [t.manage, t.workflow, workflowId];
  const id1 = useId();
  const id2 = useId();
  const id3 = useId();

  const nodes = new Map([
    ["node1", <ChannelPolicyNode id={id1} key={"node1"} nodeName={"firstNode"} />],
    ["node2", <ChannelPolicyNode id={id2} key={"node2"} nodeName={"secondNode"} />],
    ["node3", <ChannelPolicyNode id={id3} key={"node3"} nodeName={"thirdNode"} />],
  ]);

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={t.workflow}>
        {props.title}
      </PageTitle>
      <WorkflowCanvas>{Array.from(nodes.values())}</WorkflowCanvas>
    </div>
  );
}

export default WorkflowPage;
