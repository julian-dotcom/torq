import { Delete12Regular as CloseIcon, EditRegular as EditIcon } from "@fluentui/react-icons";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./workflow_cell.module.scss";
import classNames from "classnames";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useUpdateWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { Status } from "constants/backend";
import mixpanel from "mixpanel-browser";

interface WorkflowCell {
  name: string;
  workflowId: number;
  workflowVersionId: number;
  className?: string;
}

function WorkflowCell(props: WorkflowCell) {
  const { t } = useTranslations();

  const [updateWorkflow] = useUpdateWorkflowMutation();

  function archiveWorkflow() {
    if (!confirm(t.confirmDeleteWorkflow)) {
      return;
    }
    mixpanel.track("Workflow Archive", {
      workflowId: props.workflowId,
      workflowVersionId: props.workflowVersionId,
      workflowName: props.name,
    });
    updateWorkflow({ workflowId: props.workflowId, status: Status.Archived });
  }

  const content = (
    <div className={styles.workflowCellContent}>
      <div className={styles.name}>{props.name}</div>
      <div className={styles.actionButtons}>
        <LinkButton
          to={`${props.workflowId}/versions/${props.workflowVersionId}`}
          onClick={() => {
            mixpanel.track("Navigate to Workflow", {
              workflowId: props.workflowId,
              workflowVersion: props.workflowVersionId,
              workflowName: props.name,
            });
          }}
          icon={<EditIcon />}
          buttonSize={SizeVariant.tiny}
          hideMobileText={true}
        >
          {t.edit}
        </LinkButton>
        <Button
          icon={<CloseIcon />}
          buttonSize={SizeVariant.tiny}
          buttonColor={ColorVariant.error}
          onClick={archiveWorkflow}
          hideMobileText={true}
        >
          {t.archive}
        </Button>
      </div>
    </div>
  );

  return (
    <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>
      {content}
    </div>
  );
}
export default WorkflowCell;
