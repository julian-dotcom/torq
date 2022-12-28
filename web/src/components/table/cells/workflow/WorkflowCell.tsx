import { Delete12Regular as CloseIcon } from "@fluentui/react-icons"; // Edit16Regular as UpdateIcon,
import { Link } from "react-router-dom";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./workflow_cell.module.scss";
import classNames from "classnames";

interface WorkflowCell {
  name: string;
  workflowId: number;
  workflowVersionId: number;
  className?: string;
}

function WorkflowCell(props: WorkflowCell) {
  // const { t } = useTranslations();
  // const location = useLocation();

  const content = (
    <div className={styles.workflowCellContent}>
      <div className={classNames(cellStyles.current, cellStyles.text)}>{props.name}</div>
      <div className={styles.actionButtons}>
        {/*<Link to={`${props.workflowId}`} className={classNames(cellStyles.action, styles.updateLink)}>*/}
        {/*  <UpdateIcon /> {t.edit}*/}
        {/*</Link>*/}
        <div className={classNames(cellStyles.action, styles.closeChannelLink)}>
          <CloseIcon /> Close
        </div>
      </div>
    </div>
  );

  return (
    <Link to={`${props.workflowId}/versions/${props.workflowVersionId}`} className={styles.workflowCellWrapper}>
      <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>
        {content}
      </div>
    </Link>
  );
}
export default WorkflowCell;
