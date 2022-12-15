import useTranslations from "services/i18n/useTranslations";
import styles from "./workflow_page.module.scss";
import PageTitle from "features/templates/PageTitle";
import { MutableRefObject, useRef } from "react";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import ChannelPolicyNode from "components/workflow/nodes/ChannelPolicy";

type WorkflowPageProps = {
  title?: string;
  breadcrumbs?: Array<any>;
};

function WorkflowPage<T>(props: WorkflowPageProps) {
  const { t } = useTranslations();
  const bradcrumbs = props.breadcrumbs || ["Workflow", "1"];

  const canvasRef = useRef() as MutableRefObject<HTMLDivElement>;
  const blankImgRef = useRef() as MutableRefObject<HTMLCanvasElement>;

  return (
    <div className={styles.contentWrapper}>
      <PageTitle breadcrumbs={bradcrumbs} title={t.workflow}>
        {props.title}
      </PageTitle>

      <WorkflowCanvas canvasRef={canvasRef} blankImgRef={blankImgRef}>
        <ChannelPolicyNode canvasRef={canvasRef} blankImgRef={blankImgRef} />
        <canvas ref={blankImgRef} style={{ width: "1px", height: "1px" }} />
      </WorkflowCanvas>
    </div>
  );
}

export default WorkflowPage;
