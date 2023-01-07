import cellStyles from "components/table/cells/cell.module.scss";
import {
  EditRegular as UpdateIcon,
  Delete12Regular as CloseIcon,
  Eye12Regular as InspectIcon,
} from "@fluentui/react-icons";
import styles from "./channel_cell.module.scss";
import { useLocation } from "react-router-dom";
import classNames from "classnames";
import { CLOSE_CHANNEL, UPDATE_CHANNEL } from "constants/routes";
import { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";

interface ChannelCell {
  alias: string;
  channelId: number;
  nodeId: number;
  open?: boolean;
  className?: string;
}

function ChannelCell(props: ChannelCell) {
  const { t } = useTranslations();
  const location = useLocation();

  const content = (
    <>
      <div className={classNames(cellStyles.current, cellStyles.text)}>{props.alias}</div>
      <div className={styles.actionButtons}>
        <LinkButton
          key={"buttons-node-inspect"}
          state={{ background: location }}
          to={"/analyse/inspect/" + props.channelId}
          icon={<InspectIcon />}
          hideMobileText={true}
          buttonSize={SizeVariant.tiny}
          buttonColor={ColorVariant.accent1}
        >
          {t.inspect}
        </LinkButton>

        <LinkButton
          to={`${UPDATE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`}
          state={{ background: location }}
          hideMobileText={true}
          icon={<UpdateIcon />}
          className={classNames(cellStyles.action, styles.updateLink)}
          buttonSize={SizeVariant.tiny}
        >
          Update
        </LinkButton>

        <LinkButton
          to={`${CLOSE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`}
          state={{ background: location }}
          hideMobileText={true}
          className={classNames(cellStyles.action, styles.closeChannelLink)}
          icon={<CloseIcon />}
          buttonSize={SizeVariant.tiny}
        >
          Close
        </LinkButton>
      </div>
    </>
  );

  return (
    <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>
      {content}
    </div>
  );
}
export default ChannelCell;
