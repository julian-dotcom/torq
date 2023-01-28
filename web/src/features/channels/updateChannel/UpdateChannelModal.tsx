import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { useUpdateChannelMutation } from "apiSlice";
import { useState, useEffect } from "react";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./updateChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import { NumberFormatValues } from "react-number-format";
import clone from "clone";
import FormRow from "features/forms/FormWrappers";
import { useSearchParams } from "react-router-dom";
import Input from "components/forms/input/Input";

const updateStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const updateStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function NodechannelModal() {
  const { t } = useTranslations();
  const [queryParams] = useSearchParams();
  const nodeId = parseInt(queryParams.get("nodeId") || "0");
  const channelId = parseInt(queryParams.get("channelId") || "0");

  const [updateChannelMutation, response] = useUpdateChannelMutation();
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [errMessage, setErrorMessage] = useState<any[]>([]);

  useEffect(() => {
    if (response.isSuccess) {
      if (response.data.status != 1) {
        setResultState(ProgressStepState.error);
        const message = clone(errMessage) || [];
        if (response.data?.failedUpdates?.length) {
          for (let i = 0; i < response.data.failedUpdates.length; i++) {
            message.push(
              <span key={i} className={classNames(styles.updateChannelStatusMessage)}>
                {" "}
                {response.data.failedUpdates[i].reason}{" "}
              </span>
            );
          }
          setErrorMessage(message);
        }
      } else {
        setResultState(ProgressStepState.completed);
      }
    }
  }, [response]);

  const [policyState, setPolicyState] = useState(ProgressStepState.disabled);
  const [feeRateMilliMsat, setFeeRateMilliMsat] = useState<number | undefined>(undefined);
  const [feeBaseMsat, setFeeBaseMsat] = useState<number | undefined>(undefined);
  const [minHtlcMsat, setMinHtlcMsat] = useState<number | undefined>(undefined);
  const [maxHtlcMsat, setMaxHtlcMsat] = useState<number | undefined>(undefined);
  const [timeLockDelta, setTimeLockDelta] = useState<number | undefined>(undefined);
  const [stepIndex, setStepIndex] = useState(0);

  const closeAndReset = () => {
    setStepIndex(0);
    setPolicyState(ProgressStepState.active);
    setResultState(ProgressStepState.disabled);
    setErrorMessage([]);
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={"Update Channel"} show={true} onClose={() => navigate(-1)} icon={<ChannelsIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Policy"} state={policyState} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.feeRateMilliMsat}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" ppm"}
                    thousandSeparator={","}
                    value={feeRateMilliMsat}
                    onValueChange={(values: NumberFormatValues) => {
                      setFeeRateMilliMsat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.feeBaseMsat}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" msat"}
                    thousandSeparator={","}
                    value={feeBaseMsat}
                    onValueChange={(values: NumberFormatValues) => {
                      setFeeBaseMsat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
            </FormRow>

            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.minHtlcMsat}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" msat"}
                    thousandSeparator={","}
                    value={minHtlcMsat}
                    onValueChange={(values: NumberFormatValues) => {
                      setMinHtlcMsat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.maxHtlcMsat}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" msat"}
                    thousandSeparator={true}
                    value={maxHtlcMsat}
                    onValueChange={(values: NumberFormatValues) => {
                      setMaxHtlcMsat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
            </FormRow>

            <div className={styles.updateChannelTableRow}>
              <FormRow>
                <div className={styles.updateChannelTableSingle}>
                  <span className={styles.label}>{"Time Lock Delta"}</span>
                  <div className={styles.input}>
                    <Input
                      formatted={true}
                      className={styles.single}
                      thousandSeparator={false}
                      value={timeLockDelta}
                      onValueChange={(values: NumberFormatValues) => {
                        setTimeLockDelta(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <ButtonWrapper
              rightChildren={
                <Button
                  onClick={() => {
                    setStepIndex(1);
                    setPolicyState(ProgressStepState.completed);
                    setResultState(ProgressStepState.processing);
                    updateChannelMutation({
                      feeRateMilliMsat: feeRateMilliMsat,
                      feeBaseMsat: feeBaseMsat,
                      timeLockDelta: timeLockDelta,
                      minHtlcMsat: minHtlcMsat,
                      maxHtlcMsat: maxHtlcMsat,
                      channelId: channelId,
                      nodeId: nodeId,
                    });
                  }}
                  buttonColor={ColorVariant.success}
                >
                  {t.updateChannelPolicy.update}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.updateChannelResultIconWrapper,
              { [styles.failed]: !response.data },
              updateStatusClass[response.data?.status == 1 ? "SUCCEEDED" : "FAILED"]
            )}
          >
            {" "}
            {!response.data && updateStatusIcon["FAILED"]}
            {updateStatusIcon[response.data?.status == 1 ? "SUCCEEDED" : "FAILED"]}
          </div>
          <div className={errMessage.length ? styles.errorBox : styles.successeBox}>
            <div>
              <div className={errMessage.length ? styles.errorIcon : styles.successIcon}>
                {updateStatusIcon["NOTE"]}
              </div>
              <div className={errMessage.length ? styles.errorNote : styles.successNote}>
                {errMessage.length ? t.openCloseChannel.error : t.openCloseChannel.note}
              </div>
            </div>
            <div className={errMessage.length ? styles.errorMessage : styles.successMessage}>
              {errMessage.length ? errMessage : t.updateChannelPolicy.confirmedMessage}
            </div>
          </div>
          <ButtonWrapper
            rightChildren={
              <Button
                onClick={() => {
                  setStepIndex(0);
                  setPolicyState(ProgressStepState.active);
                  setResultState(ProgressStepState.disabled);
                  setErrorMessage([]);
                }}
                buttonColor={ColorVariant.primary}
              >
                {t.updateChannelPolicy.newUpdate}
              </Button>
            }
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NodechannelModal;
