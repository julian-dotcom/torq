import { Delete12Regular as CloseIcon, Edit16Regular as EditIcon } from "@fluentui/react-icons"; // Edit16Regular as UpdateIcon,
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./workflow_cell.module.scss";
import classNames from "classnames";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useUpdateWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { Status } from "../../../../constants/backend";

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
    updateWorkflow({ workflowId: props.workflowId, status: Status.Archived });
  }

  // function handleWorkflowNameChange(name: string) {
  //   updateWorkflow({ workflowId: props.workflowId, name: name });
  // }

  const content = (
    <div className={styles.workflowCellContent}>
      <div className={styles.name}>{props.name}</div>
      <div className={styles.actionButtons}>
        <LinkButton
          to={`${props.workflowId}/versions/${props.workflowVersionId}`}
          icon={<EditIcon />}
          buttonSize={SizeVariant.small}
        >
          {t.edit}
        </LinkButton>
        <Button
          icon={<CloseIcon />}
          buttonSize={SizeVariant.small}
          buttonColor={ColorVariant.error}
          onClick={archiveWorkflow}
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
