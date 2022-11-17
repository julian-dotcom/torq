import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowRouting20Regular as ChannelsIcon,
  CommentLightning20Regular as AdvencedOption,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, WS_URL, useGetChannelsQuery } from "apiSlice";
import type { channel, nodeConfiguration } from "apiTypes";
import { useState, useEffect, ChangeEvent } from "react";
import Button, { buttonColor, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./closeChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import classNames from "classnames";
import NumberFormat, { NumberFormatValues } from "react-number-format";
import Input from "components/forms/input/Input";
import { SectionContainer } from "features/section/SectionContainer";
import useWebSocket from "react-use-websocket";
import Switch from "components/forms/switch/Switch";
import FormRow from "features/forms/FormWrappers";

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
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const { data: channels } = useGetChannelsQuery();

  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<SelectOptions[]>([{ value: 0, label: "Select a local node" }]);
  const [channelOptions, setChannelOptions] = useState<SelectOptions[]>([{ value: 0, label: "Select your channel" }]);

  useEffect(() => {
    if (channels !== undefined) {
      const newChannelOptions = channels.map((channel: channel) => {
        return {
          value: channel.channelPoint,
          label: `${channel.peerAlias} - ${channel.lndShortChannelId.toString()}`,
        };
      });
      setChannelOptions(newChannelOptions);
    }
    if (nodeConfigurations !== undefined) {
        const newNodeOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
          return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
        });
        setNodeConfigurationOptions(newNodeOptions);
      }
  }, [channels, nodeConfigurations]);

  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value as number);
  const [selectedChannel, setSelectedChannel] = useState<string>(channelOptions[0]?.value as string);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [errMessage, setErrorMEssage] = useState<string>("");
  const [detailState, setDetailState] = useState(ProgressStepState.active);
  const [closeAddress, setCloseAddress] = useState<string>("");
  const [satPerVbyte, setSatPerVbyte] = useState<number>(0);
  const [stepIndex, setStepIndex] = useState(0);
  const [force, setForce] = useState<boolean>(false);

  function handleNodeSelection(value: number) {
    setSelectedNodeId(value);
    const filteredChannels = channels?.filter((channel: { nodeId: number }) => channel.nodeId == value);
      const filteredChannelOptions = filteredChannels?.map((channel: channel) => {
        if (channel.nodeId == value) {
          return {
            value: channel.channelPoint,
            label: `${channel.peerAlias} - ${channel.lndShortChannelId.toString()}`,
          };
        }
      });
      setChannelOptions(filteredChannelOptions as SelectOptions[]);
  }

  function handleChannelSelection(value: string) {
    setSelectedChannel(value);
  }

  const closeAndReset = () => {
    setStepIndex(0);
    setSelectedNodeId(0);
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
          <Select
            label={t.yourNode}
            onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
              const selectOptions = newValue as SelectOptions;
              handleNodeSelection(selectOptions?.value as number);
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
          <Select
            label={t.yourChannel}
            onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
              const selectOptions = newValue as SelectOptions;
              handleChannelSelection(selectOptions?.value as string);
            }}
            options={channelOptions}
            value={channelOptions.find((option) => option.value === selectedChannel)}
            isDisabled={true}
          />
          <div className={styles.activeColumns}>
            <SectionContainer
            title={"Advanced Options"}
            icon={AdvencedOption}
            expanded={expandAdvancedOptions}
            handleToggle={() => {
              setExpandAdvancedOptions(!expandAdvancedOptions);
            }}
          >
            <div className={styles.closeChannelTableRow}>
              <FormRow>
                <div className={styles.closeChannelTableSingle}>
                  <span className={styles.label}>{"Sat per vbyte"}</span>
                  <div className={styles.input}>
                    <NumberFormat
                      className={styles.single}
                      thousandSeparator={false}
                      value={satPerVbyte}
                      onValueChange={(values: NumberFormatValues) => {
                        setSatPerVbyte(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.closeChannelTableRow}>
              <FormRow>
                <div className={styles.closeChannelTableSingle}>
                  <span className={styles.label}>{"Close Address (for local funds)"}</span>
                  <div className={styles.input}>
                    <Input
                      type={"text"}
                      value={closeAddress}
                      placeholder={"e.g. bc1q..."}
                      onChange={(e: ChangeEvent<HTMLInputElement>) => {
                        setCloseAddress(e.target.value);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.closeChannelTableRow}>
              <FormRow className={styles.switchRow}>
              <Switch label={"Force close"}
                checked={force}
                onChange={(e: ChangeEvent<HTMLInputElement>) => {
                  setForce(e.target.checked)
                }}
              />
              </FormRow>
            </div>
          </SectionContainer>
            <ButtonWrapper
              rightChildren={
                <Button
                  text={t.openCloseChannel.closeChannel}
                  onClick={() => {
                    setStepIndex(1);
                    setDetailState(ProgressStepState.completed);
                    setResultState(ProgressStepState.completed);
                    sendJsonMessage({
                      reqId: "randId",
                      type: "closeChannel",
                      closeChannelRequest: {
                        nodeId: selectedNodeId,
                        channelpoint: selectedChannel,
                        satPerVbyte,
                        deliveryAddress: closeAddress,
                        force,
                      },
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
              styles.closeChannelResultIconWrapper,
              { [styles.failed]: errMessage },
              closeStatusClass[errMessage ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {" "}
            {closeStatusIcon[errMessage ? "FAILED" : "SUCCEEDED"]}
          </div>
          <div className={errMessage ? styles.errorBox : styles.successeBox }>
            <div>
              <div className={errMessage ? styles.errorIcon : styles.successIcon }>{closeStatusIcon["NOTE"]}</div>
              <div className={errMessage ? styles.errorNote : styles.successNote}>{errMessage ? t.openCloseChannel.error :t.openCloseChannel.note}</div>
            </div >
            <div className={errMessage ? styles.errorMessage: styles.successMessage }>
              {errMessage ? errMessage : t.openCloseChannel.confirmationClosing}
            </div>
          </div>
          <ButtonWrapper
            rightChildren={
              <Button
                text={t.openCloseChannel.closeNewChannel}
                onClick={() => {
                  setStepIndex(0);
                  setDetailState(ProgressStepState.active);
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

export default closeChannelModal;
