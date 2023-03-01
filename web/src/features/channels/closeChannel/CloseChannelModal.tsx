import {
  ArrowRouting20Regular as ChannelsIcon,
  ArrowSyncFilled as ProcessingIcon,
  Checkmark20Regular as SuccessNoteIcon,
  CheckmarkRegular as SuccessIcon,
  ErrorCircleRegular as FailedIcon,
  ErrorCircle20Regular as FailedNoteIcon,
  Link20Regular as LinkIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { WS_URL } from "apiSlice";
import { ChangeEvent, useState } from "react";
import Button, { ButtonWrapper, ColorVariant, ExternalLinkButton } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./closeChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import { NumberFormatValues } from "react-number-format";
import Input from "components/forms/input/Input";
import useWebSocket from "react-use-websocket";
import Switch from "components/forms/switch/Switch";
import FormRow from "features/forms/FormWrappers";
import { useSearchParams } from "react-router-dom";
import { Buffer } from "buffer";
import Note, { NoteType } from "features/note/Note";
import mixpanel from "mixpanel-browser";
import { FormErrors, mergeServerError } from "components/errors/errors";

const closeStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const closeStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function closeChannelModal() {
  const { t } = useTranslations();
  const [queryParams] = useSearchParams();
  const nodeId = parseInt(queryParams.get("nodeId") || "0");
  const channelId = parseInt(queryParams.get("channelId") || "0");

  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [errMessage, setErrorMessage] = useState<string>("");
  const [closingTx, setClosingTx] = useState<string>("");
  const [detailState, setDetailState] = useState(ProgressStepState.active);
  const [satPerVbyte, setSatPerVbyte] = useState<number | undefined>();
  const [stepIndex, setStepIndex] = useState(0);
  const [force, setForce] = useState<boolean>(false);
  const [formErrorState, setFormErrorState] = useState({} as FormErrors);

  const closeAndReset = () => {
    setStepIndex(0);
    setDetailState(ProgressStepState.active);
    setResultState(ProgressStepState.disabled);
    setErrorMessage("");
    setClosingTx("");
  };

  const navigate = useNavigate();

  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: oncloseChannelMessage,
  });

  function oncloseChannelMessage(event: MessageEvent<string>) {
    setStepIndex(1);
    const response = JSON.parse(event.data);
    if (response?.type === "Error") {
      setFormErrorState(mergeServerError(response.error, formErrorState));
      setErrorMessage(response.error);
      setResultState(ProgressStepState.error);
      return;
    } else {
      if (!closingTx) {
        const decodedTxId = Buffer.from(response.closePendingChannelPoint.txId, "base64").toString("utf8");
        setClosingTx(`${decodedTxId}`);
      }
    }
  }

  function closeChannel() {
    setDetailState(ProgressStepState.processing);
    setResultState(ProgressStepState.completed);
    mixpanel.track("Close Channel", {
      nodeId: nodeId,
      channelId: channelId,
      openChannelUseSatPerVbyte: satPerVbyte !== 0,
      force: force,
    });
    sendJsonMessage({
      requestId: "randId",
      type: "closeChannel",
      closeChannelRequest: {
        nodeId: nodeId,
        channelId: channelId,
        satPerVbyte: satPerVbyte,
        force: force,
      },
    });
  }
  return (
    <PopoutPageTemplate title={"Close Channel"} show={true} onClose={() => navigate(-1)} icon={<ChannelsIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Detail"} state={detailState} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <div className={styles.closeChannelTableRow}>
              <FormRow>
                <div className={styles.closeChannelTableSingle}>
                  <span className={styles.label}>{"Sat per vbyte"}</span>
                  <div className={styles.input}>
                    <Input
                      formatted={true}
                      className={styles.single}
                      thousandSeparator={","}
                      value={satPerVbyte}
                      suffix={" sat/vbyte"}
                      onValueChange={(values: NumberFormatValues) => {
                        setSatPerVbyte(values.floatValue);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.closeChannelTableRow}>
              <FormRow className={styles.switchRow}>
                <Switch
                  label={"Force close"}
                  checked={force}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => {
                    setForce(e.target.checked);
                  }}
                />
              </FormRow>
            </div>
            <ButtonWrapper
              rightChildren={
                <Button onClick={closeChannel} buttonColor={ColorVariant.success}>
                  {t.openCloseChannel.closeChannel}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.closeChannelResultIconWrapper,
              { [styles.failed]: errMessage },
              closeStatusClass[errMessage ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {" "}
            {closeStatusIcon[errMessage ? "FAILED" : "SUCCEEDED"]}
          </div>
          <div className={styles.closeChannelResultDetails}>
            {!errMessage && (
              <>
                <Note title={t.TxId} icon={<SuccessNoteIcon />} noteType={NoteType.success}>
                  {closingTx}
                </Note>
                <ExternalLinkButton
                  href={"https://mempool.space/tx/" + closingTx}
                  target="_blank"
                  rel="noreferrer"
                  buttonColor={ColorVariant.success}
                  icon={<LinkIcon />}
                >
                  {t.openCloseChannel.GoToMempool}
                </ExternalLinkButton>

                <Note title={t.note} icon={<NoteIcon />} noteType={NoteType.info}>
                  {t.openCloseChannel.confirmationClosing}
                </Note>
              </>
            )}
            {errMessage && (
              <Note title={t.openCloseChannel.error} icon={<FailedNoteIcon />} noteType={NoteType.error}>
                {errMessage}
              </Note>
            )}
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default closeChannelModal;
