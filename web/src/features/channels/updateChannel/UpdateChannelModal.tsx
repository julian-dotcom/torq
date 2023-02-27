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
import mixpanel from "mixpanel-browser";
import { PolicyInterface } from "features/channels/channelsTypes";
import ErrorSummary from "components/errors/ErrorSummary";
import { FormErrors, mergeServerError, ServerErrorType } from "components/errors/errors";
import Note from "features/note/Note";

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

function UpdateChannelModal() {
  const { t } = useTranslations();
  const [queryParams] = useSearchParams();
  const nodeId = parseInt(queryParams.get("nodeId") || "0");
  const channelId = parseInt(queryParams.get("channelId") || "0");

  const [updateChannelMutation, response] = useUpdateChannelMutation();
  const [policyState, setPolicyState] = useState(ProgressStepState.active);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [formErrorState, setFormErrorState] = useState({} as FormErrors);
  const [feeRateMilliMsat, setFeeRateMilliMsat] = useState<number | undefined>(undefined);
  const [timeLockDelta, setTimeLockDelta] = useState<number | undefined>(undefined);
  const [feeBase, setFeeBase] = useState<number | undefined>(undefined);
  const [maxHtlc, setMaxHtlc] = useState<number | undefined>(undefined);
  const [minHtlc, setMinHtlc] = useState<number | undefined>(undefined);
  const [stepIndex, setStepIndex] = useState(0);

  useEffect(() => {
    if (response && response.error && "data" in response.error && response.error.data) {
      const mergedErrors = mergeServerError(response.error.data as ServerErrorType, clone(formErrorState));
      setFormErrorState(mergedErrors);
      setResultState(ProgressStepState.error);
      return;
    }
    if (response.isSuccess) {
      setResultState(ProgressStepState.completed);
    }
  }, [response]);

  const closeAndReset = () => {
    setStepIndex(0);
    setPolicyState(ProgressStepState.active);
    setResultState(ProgressStepState.disabled);
    setFormErrorState({} as FormErrors);
  };

  const navigate = useNavigate();

  const updateChannel = () => {
    setStepIndex(1);
    setPolicyState(ProgressStepState.completed);
    setResultState(ProgressStepState.processing);
    const pi: PolicyInterface = {
      feeRateMilliMsat: feeRateMilliMsat,
      timeLockDelta: timeLockDelta,
      channelId: channelId,
      nodeId: nodeId,
    };
    mixpanel.track("Update Channel", {
      channelId: channelId,
      nodeId: nodeId,
    });

    if (feeBase !== undefined) {
      pi.feeBaseMsat = feeBase * 1000;
    }
    if (maxHtlc !== undefined) {
      pi.maxHtlcMsat = maxHtlc * 1000;
    }
    if (minHtlc !== undefined) {
      pi.minHtlcMsat = minHtlc * 1000;
    }
    updateChannelMutation(pi);
  };

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
                <span className={styles.label}>{t.updateChannelPolicy.feeBase}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" sat"}
                    thousandSeparator={","}
                    value={feeBase}
                    onValueChange={(values: NumberFormatValues) => {
                      setFeeBase(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
            </FormRow>

            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.minHtlc}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" sat"}
                    thousandSeparator={","}
                    value={minHtlc}
                    onValueChange={(values: NumberFormatValues) => {
                      setMinHtlc(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.maxHtlc}</span>
                <div className={styles.input}>
                  <Input
                    formatted={true}
                    className={styles.double}
                    suffix={" sat"}
                    thousandSeparator={true}
                    value={maxHtlc}
                    onValueChange={(values: NumberFormatValues) => {
                      setMaxHtlc(values.floatValue as number);
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
                <Button onClick={updateChannel} buttonColor={ColorVariant.success}>
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
            {updateStatusIcon[response.data?.status == 1 ? "SUCCEEDED" : "FAILED"]}
          </div>
          <ErrorSummary errors={formErrorState} />
          {resultState === ProgressStepState.completed && (
            <Note title="Success">{t.updateChannelPolicy.confirmedMessage}</Note>
          )}
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default UpdateChannelModal;
