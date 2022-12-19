import cellStyles from "components/table/cells/cell.module.scss";
// import { Edit16Regular as UpdateIcon } from "@fluentui/react-icons";
import styles from "./channel_cell.module.scss";
import { Link, useLocation } from "react-router-dom";
import classNames from "classnames";
import useTranslations from "services/i18n/useTranslations";

interface WorkflowCell {
  name: string;
  workflowId: number;
  className?: string;
}

function WorkflowCell(props: WorkflowCell) {
  const { t } = useTranslations();
  const location = useLocation();

  const content = (
    <>
      <div className={classNames(cellStyles.current, cellStyles.text)}>{props.name}</div>
      <div className={styles.actionButtons}>
        close
        {/*<Link to={`${props.workflowId}`} className={classNames(cellStyles.action, styles.updateLink)}>*/}
        {/*  <UpdateIcon /> {t.edit}*/}
        {/*</Link>*/}
        {/*<Link*/}
        {/*  to={`${}?nodeId=${props.nodeId}&channelId=${props.channelId}`}*/}
        {/*  state={{ background: location }}*/}
        {/*  className={classNames(cellStyles.action, styles.closeChannelLink)}*/}
        {/*>*/}
        {/*  <CloseIcon /> Close*/}
        {/*</Link>*/}
      </div>
    </>
  );

  return (
    <Link to={`${props.workflowId}`}>
      <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>
        {content}
      </div>
    </Link>
  );
}
export default WorkflowCell;
