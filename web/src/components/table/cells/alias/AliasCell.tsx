import styles from "components/table/cells/cell.module.scss";
import {
  Delete12Regular as CloseIcon,
  EditRegular as EditIcon,
  Eye12Regular as InspectIcon,
} from "@fluentui/react-icons";
import classNames from "classnames";
import { ButtonPosition, ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { CLOSE_CHANNEL, UPDATE_CHANNEL } from "constants/routes";
import { useLocation } from "react-router-dom";
import mixpanel from "mixpanel-browser";

interface AliasCell {
  current: string;
  channelId: number;
  nodeIds: Array<number>;
  open?: boolean;
  className?: string;
  isTotalsRow?: boolean;
}

function AliasCell({ current, nodeIds, channelId, open, className, isTotalsRow }: AliasCell) {
  const { t } = useTranslations();
  const location = useLocation();
  const content = (
    <div className={styles.alias}>
      <div className={classNames(styles.current, styles.text)}>{current}</div>
      {open !== undefined && (
        <div className={classNames(styles.buttonWrapper, { [styles.totalCell]: isTotalsRow })}>
          <LinkButton
            key={"buttons-node-inspect"}
            state={{ background: location }}
            to={"/analyse/inspect/" + channelId}
            icon={<InspectIcon />}
            hideMobileText={true}
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.accent1}
            buttonPosition={ButtonPosition.center}
            onClick={() => {
              mixpanel.track("Navigate to Inspect Channel", {
                background: location.pathname,
                channelId: channelId,
              });
            }}
          >
            {t.inspect}
          </LinkButton>
          {(nodeIds || []).map((nodeId) => {
            return (
              <div className={styles.editChannelButton} key={"buttons-node-" + nodeId}>
                <LinkButton
                  to={`${UPDATE_CHANNEL}?nodeId=${nodeId}&channelId=${channelId}`}
                  state={{ background: location }}
                  className={classNames(styles.action, styles.updateLink)}
                  buttonSize={SizeVariant.tiny}
                  buttonColor={ColorVariant.success}
                  hideMobileText={true}
                  icon={<EditIcon />}
                  onClick={() => {
                    mixpanel.track("Navigate to Update Channel", {
                      background: location.pathname,
                      channelId: channelId,
                    });
                  }}
                >
                  {t.update}
                </LinkButton>

                <LinkButton
                  to={`${CLOSE_CHANNEL}?nodeId=${nodeId}&channelId=${channelId}`}
                  state={{ background: location }}
                  className={classNames(styles.action, styles.closeChannelLink)}
                  buttonSize={SizeVariant.tiny}
                  buttonColor={ColorVariant.error}
                  hideMobileText={true}
                  icon={<CloseIcon />}
                  onClick={() => {
                    mixpanel.track("Navigate to Close Channel", {
                      background: location.pathname,
                      channelId: channelId,
                    });
                  }}
                >
                  {t.close}
                </LinkButton>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );

  return <div className={classNames(styles.cell, styles.alignLeft, className)}>{content}</div>;
}
export default AliasCell;
