import cellStyles from "components/table/cells/cell.module.scss";
import { Edit16Regular as UpdateIcon, Delete12Regular as CloseIcon } from "@fluentui/react-icons";
import styles from "./channel_cell.module.scss";
import {Link, useLocation} from "react-router-dom";
import classNames from "classnames";
import {CLOSE_CHANNEL, UPDATE_CHANNEL} from "constants/routes";

interface ChannelCell {
  alias: string;
  channelId: number;
  nodeId: number;
  open?: boolean;
  className?: string;
}

function ChannelCell(props: ChannelCell) {

  const location = useLocation();

  const content = (
    <>
      <div className={classNames(cellStyles.current, cellStyles.text)}>{props.alias}</div>
      <div className={styles.actionButtons}>
        <Link to={`${UPDATE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`} state={{ background: location }} className={classNames(cellStyles.action, styles.updateLink)}>
          <UpdateIcon  /> Update
        </Link>

        <Link to={`${CLOSE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`} state={{ background: location }} className={classNames(cellStyles.action, styles.closeChannelLink)}>
          <CloseIcon  /> Close
        </Link>
      </div>
    </>
  )

  return <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>{content}</div>
}
export default ChannelCell;
