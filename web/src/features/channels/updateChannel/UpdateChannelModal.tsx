import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowRouting20Regular as ChannelsIcon,
} from "@fluentui/react-icons";
import { useGetLocalNodesQuery, useGetChannelsQuery, useUpdateChannelMutation } from "apiSlice";
import type { channel } from "apiTypes";
import { useState, useEffect } from "react";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./updateChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import { localNode } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import classNames from "classnames";
import NumberFormat, { NumberFormatValues } from "react-number-format";

import clone from "clone";
import FormRow from "../../forms/FormWrappers";

const updateStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const updateStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
};

function NodechannelModal() {
  const [updateChannelMutation, response] = useUpdateChannelMutation();

  const { t } = useTranslations();

  const { data: localNodes } = useGetLocalNodesQuery();
  let { data: channels } = useGetChannelsQuery();

  let localNodeOptions: SelectOptions[] = [{ value: 0, label: "Select a local node" }];
  if (localNodes !== undefined) {
    localNodeOptions = localNodes.map((localNode: localNode) => {
      return { value: localNode.localNodeId, label: localNode.name };
    });
  }
  let channelOptions: SelectOptions[] = [{ value: 0, label: "Select your channel" }];
  if (channels !== undefined) {
    channelOptions = channels.map((channel: channel) => {
      return {
        value: channel.lndShortChannelId,
        label: `${channel.peerAlias} - ${channel.lndShortChannelId.toString()}`,
      };
    });
  }

  const [selectedLocalNode, setSelectedLocalNode] = useState<number>(localNodeOptions[0].value);
  const [selectedChannel, setSelectedChannel] = useState<number>(channelOptions[0].value);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [errMessage, setErrorMEssage] = useState<any[]>([]);

  function handleNodeSelection(value: number) {
    setSelectedLocalNode(value);
    channels = channels?.filter((channel: { localNodeId: number }) => channel.localNodeId == value);
  }

  function handleChannelSelection(value: number) {
    setSelectedChannel(value);
    channels?.map((channel: channel) => {
      if (channel.lndShortChannelId == value) {
        setTimeLockDelta(channel.timeLockDelta);
        setBaseFeeMsat(channel.baseFeeMsat);
        setMinHtlcSat(channel.minHtlc / 1000);
        setMaxHtlcSat(channel.maxHtlcMsat / 1000);
        setFeeRatePpm(channel.feeRatePpm);
        setLndChannelPoint(channel.lndChannelPoint);
        return channel;
      }
    });
  }

  useEffect(() => {
    if (response.isSuccess) {
      if (response.data.status == "FAILED") {
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
          setErrorMEssage(message);
        }
      } else {
        setResultState(ProgressStepState.completed);
      }
    }
  }, [response]);

  const [channelState, setChannelState] = useState(ProgressStepState.active);
  const [policyState, setPolicyState] = useState(ProgressStepState.disabled);
  const [feeRatePpm, setFeeRatePpm] = useState<number>(0);
  const [baseFeeMsat, setBaseFeeMsat] = useState<number>(0);
  const [minHtlcSat, setMinHtlcSat] = useState<number>(0);
  const [maxHtlcSat, setMaxHtlcSat] = useState<number>(0);
  const [timeLockDelta, setTimeLockDelta] = useState<number>(0);
  const [lndChannelPoint, setLndChannelPoint] = useState<string>("");
  const [stepIndex, setStepIndex] = useState(0);

  const closeAndReset = () => {
    setStepIndex(0);
    setSelectedLocalNode(0);
    setSelectedChannel(0);
    setChannelState(ProgressStepState.active);
    setPolicyState(ProgressStepState.disabled);
    setResultState(ProgressStepState.disabled);
    setErrorMEssage([]);
  };

  const dynamicChannelState = () => {
    if (!channels?.length) {
      return ProgressStepState.disabled;
    }
    return channelState;
  };

  const dynamicPolicyState = () => {
    if (!channels?.length) {
      return ProgressStepState.disabled;
    }
    return policyState;
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={"Update Channel"} show={true} onClose={() => navigate(-1)} icon={<ChannelsIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Channel"} state={dynamicChannelState()} last={false} />
        <Step label={"Policy"} state={dynamicPolicyState()} last={false} />
        <Step label={"Result"} state={resultState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <Select
            label={t.yourNode}
            onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
              const selectOptions = newValue as SelectOptions;
              handleNodeSelection(selectOptions?.value);
            }}
            options={localNodeOptions}
            value={localNodeOptions.find((option) => option.value === selectedLocalNode)}
          />
          <Select
            label={t.yourChannel}
            onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
              const selectOptions = newValue as SelectOptions;
              handleChannelSelection(selectOptions?.value);
            }}
            options={channelOptions}
            value={channelOptions.find((option) => option.value === selectedChannel)}
            isDisabled={true}
          />
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={"Next"}
                disabled={selectedLocalNode == 0 || selectedChannel == 0}
                onClick={() => {
                  if (selectedChannel) {
                    setStepIndex(1);
                    setChannelState(ProgressStepState.completed);
                    setPolicyState(ProgressStepState.active);
                  }
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.feeRatePpm}</span>
                <div className={styles.input}>
                  <NumberFormat
                    className={styles.double}
                    suffix={" ppm"}
                    thousandSeparator={false}
                    value={feeRatePpm}
                    onValueChange={(values: NumberFormatValues) => {
                      setFeeRatePpm(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.baseFeeMsat}</span>
                <div className={styles.input}>
                  <NumberFormat
                    className={styles.double}
                    suffix={" milli sat"}
                    thousandSeparator={false}
                    value={baseFeeMsat}
                    onValueChange={(values: NumberFormatValues) => {
                      setBaseFeeMsat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
            </FormRow>

            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.minHtlcSat}</span>
                <div className={styles.input}>
                  <NumberFormat
                    className={styles.double}
                    suffix={" sat"}
                    thousandSeparator={false}
                    value={minHtlcSat}
                    onValueChange={(values: NumberFormatValues) => {
                      setMinHtlcSat(values.floatValue as number);
                    }}
                  />
                </div>
              </div>
              <div className={styles.updateChannelTableDouble}>
                <span className={styles.label}>{t.updateChannelPolicy.maxHtlcSat}</span>
                <div className={styles.input}>
                  <NumberFormat
                    className={styles.double}
                    suffix={" sat"}
                    thousandSeparator={true}
                    value={maxHtlcSat}
                    onValueChange={(values: NumberFormatValues) => {
                      setMaxHtlcSat(values.floatValue as number);
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
                    <NumberFormat
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
                  text={t.updateChannelPolicy.update}
                  onClick={() => {
                    setStepIndex(2);
                    setPolicyState(ProgressStepState.completed);
                    setResultState(ProgressStepState.processing);
                    updateChannelMutation({
                      feeRatePpm,
                      baseFeeMsat: baseFeeMsat,
                      timeLockDelta,
                      minHtlcMsat: minHtlcSat * 1000,
                      maxHtlcMsat: maxHtlcSat * 1000,
                      channelPoint: lndChannelPoint,
                      nodeId: selectedLocalNode,
                    });
                  }}
                  buttonColor={buttonColor.green}
                />
              }
            />
          </div>
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.updateChannelResultIconWrapper,
              { [styles.failed]: !response.data },
              updateStatusClass[response.data?.status as "SUCCEEDED" | "FAILED" | "IN_FLIGHT"]
            )}
          >
            {" "}
            {!response.data && updateStatusIcon["FAILED"]}
            {updateStatusIcon[response.data?.status as "SUCCEEDED" | "FAILED" | "IN_FLIGHT"]}
          </div>
          <div className="pop">{errMessage}</div>
          <ButtonWrapper
            rightChildren={
              <Button
                text={t.updateChannelPolicy.newUpdate}
                onClick={() => {
                  setStepIndex(0);
                  setChannelState(ProgressStepState.active);
                  setPolicyState(ProgressStepState.disabled);
                  setResultState(ProgressStepState.disabled);
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NodechannelModal;
