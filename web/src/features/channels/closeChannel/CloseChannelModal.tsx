import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { WS_URL } from "apiSlice";
import { useState, ChangeEvent } from "react";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
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
  const [errMessage, setErrorMEssage] = useState<string>("");
  const [detailState, setDetailState] = useState(ProgressStepState.active);
  const [satPerVbyte, setSatPerVbyte] = useState<number | undefined>();
  const [stepIndex, setStepIndex] = useState(0);
  const [force, setForce] = useState<boolean>(false);

  const closeAndReset = () => {
    setStepIndex(0);
    setDetailState(ProgressStepState.active);
    setResultState(ProgressStepState.disabled);
    setErrorMEssage("");
  };

  const navigate = useNavigate();

  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: oncloseChannelMessage,
  });

  function oncloseChannelMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    if (response?.type === "Error") {
      setErrorMEssage(response.error);
      setResultState(ProgressStepState.error);
      return;
    }
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
            {/*<div className={styles.closeChannelTableRow}>*/}
            {/*  <FormRow>*/}
            {/*    <div className={styles.closeChannelTableSingle}>*/}
            {/*      <span className={styles.label}>{"Close Address (for local funds)"}</span>*/}
            {/*      <div className={styles.input}>*/}
            {/*        <Input*/}
            {/*          type={"text"}*/}
            {/*          value={closeAddress}*/}
            {/*          placeholder={"e.g. bc1q..."}*/}
            {/*          onChange={(e: ChangeEvent<HTMLInputElement>) => {*/}
            {/*            setCloseAddress(e.target.value);*/}
            {/*          }}*/}
            {/*        />*/}
            {/*      </div>*/}
            {/*    </div>*/}
            {/*  </FormRow>*/}
            {/*</div>*/}
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
                <Button
                  onClick={() => {
                    setStepIndex(1);
                    setDetailState(ProgressStepState.completed);
                    setResultState(ProgressStepState.completed);
                    sendJsonMessage({
                      requestId: "randId",
                      type: "closeChannel",
                      closeChannelRequest: {
                        nodeId: nodeId,
                        channelId: channelId,
                        satPerVbyte: satPerVbyte,
                        // deliveryAddress: closeAddress,
                        force: force,
                      },
                    });
                  }}
                  buttonColor={ColorVariant.success}
                >
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
          <div className={errMessage ? styles.errorBox : styles.successeBox}>
            <div>
              <div className={errMessage ? styles.errorIcon : styles.successIcon}>{closeStatusIcon["NOTE"]}</div>
              <div className={errMessage ? styles.errorNote : styles.successNote}>
                {errMessage ? t.openCloseChannel.error : t.openCloseChannel.note}
              </div>
            </div>
            <div className={errMessage ? styles.errorMessage : styles.successMessage}>
              {errMessage ? errMessage : t.openCloseChannel.confirmationClosing}
            </div>
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default closeChannelModal;
