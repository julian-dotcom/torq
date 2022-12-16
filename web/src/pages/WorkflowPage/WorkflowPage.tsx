import { useId } from "react";
import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import ChannelPolicyNode from "components/workflow/nodes/ChannelPolicy";

type WorkflowPageProps = {
  title?: string;
  breadcrumbs?: Array<any>;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const bradcrumbs = props.breadcrumbs || ["Workflow", "1"];
  const id1 = useId();
  const id2 = useId();

  const nodes = new Map([
    ["node2", <ChannelPolicyNode id={id1} key={"node2"} />],
    ["node1", <ChannelPolicyNode id={id2} key={"node1"} />],
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
