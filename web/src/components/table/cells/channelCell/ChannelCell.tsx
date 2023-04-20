import {
  EditRegular as UpdateIcon,
  Delete12Regular as CloseIcon,
  Eye12Regular as InspectIcon,
} from "@fluentui/react-icons";
import styles from "./channel_cell.module.scss";
import cellStyles from "components/table/cells/cell.module.scss";
import { useLocation } from "react-router-dom";
import classNames from "classnames";
import { CLOSE_CHANNEL, UPDATE_CHANNEL } from "constants/routes";
import { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { userEvents } from "utils/userEvents";

interface ChannelCell {
  alias: string;
  channelId: number;
  nodeId: number;
  open?: boolean;
  className?: string;
  hideActionButtons: boolean;
}

function ChannelCell(props: ChannelCell) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const location = useLocation();

  const content = (
    <>
      <div className={classNames(cellStyles.current, cellStyles.text)}>{props.alias}</div>
      {!props.hideActionButtons && (
        <div className={styles.actionButtons}>
          <LinkButton
            key={"buttons-node-inspect"}
            state={{ background: location }}
            to={"/analyse/inspect/" + props.channelId}
            icon={<InspectIcon />}
            hideMobileText={true}
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.accent1}
            onClick={() => {
              track("Navigate to Inspect Channel", {
                channelId: props.channelId,
              });
            }}
          >
            {t.inspect}
          </LinkButton>

          <LinkButton
            to={`${UPDATE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`}
            state={{ background: location }}
            hideMobileText={true}
            icon={<UpdateIcon />}
            buttonColor={ColorVariant.success}
            buttonSize={SizeVariant.tiny}
            onClick={() => {
              track("Navigate to Update Channel", {
                nodeId: props.nodeId,
                channelId: props.channelId,
              });
            }}
          >
            {t.update}
          </LinkButton>

          <LinkButton
            to={`${CLOSE_CHANNEL}?nodeId=${props.nodeId}&channelId=${props.channelId}`}
            state={{ background: location }}
            hideMobileText={true}
            icon={<CloseIcon />}
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.error}
            onClick={() => {
              track("Navigate to Close Channel", {
                nodeId: props.nodeId,
                channelId: props.channelId,
              });
            }}
          >
            {t.close}
          </LinkButton>
        </div>
      )}
    </>
  );

  return (
    <div className={classNames(cellStyles.cell, cellStyles.alignLeft, props.className, styles.channelCellWrapper)}>
      {content}
    </div>
  );
}
export default ChannelCell;
