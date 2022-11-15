import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowRouting20Regular as ChannelsIcon,
  CommentLightning20Regular as AdvencedOption,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, WS_URL } from "apiSlice";
import { useState } from "react";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./openChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import classNames from "classnames";
import NumberFormat, { NumberFormatValues } from "react-number-format";
import TextInput from "features/forms/TextInput";
import { SectionContainer } from "features/section/SectionContainer";
import useWebSocket from "react-use-websocket";
import Switch from "features/inputs/Slider/Switch";

// import clone from "clone";
import FormRow from "features/forms/FormWrappers";

const openStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const openStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function OpenChannelModal() {

  const { t } = useTranslations();
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  let nodeConfigurationOptions: SelectOptions[] = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations !== undefined) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
    });
  }

  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [errMessage, setErrorMEssage] = useState<string>("");

  function handleNodeSelection(value: number) {
    setSelectedNodeId(value);
  }

  const [connectState, setConnectState] = useState(ProgressStepState.active);
  const [detailState, setDetailState] = useState(ProgressStepState.disabled);
  const [minConfs, setMinConfs] = useState<number>(0);
  const [localFundingAmount, setLocalFundingAmount] = useState<number>(0);
  const [pushSat, setPushSat] = useState<number>(0);
  const [minHtlcMsat, setMinHtlcMsat] = useState<number>(0);
  const [closeAddress, setCloseAddress] = useState<string>("");
  const [spendUnconfirmed, setSpendUnconfirmed] = useState<boolean>(false);
  const [privateChan, setPrivate] = useState<boolean>(false);
  const [satPerVbyte, setSatPerVbyte] = useState<number>(0);
  const [nodePubKey, setNodePubKey] = useState<string>("");
  const [host, setHost] = useState<string>("");
  const [stepIndex, setStepIndex] = useState(0);

  const closeAndReset = () => {
    setStepIndex(0);
    setSelectedNodeId(0);
    setConnectState(ProgressStepState.active);
    setDetailState(ProgressStepState.disabled);
    setResultState(ProgressStepState.disabled);
    setErrorMEssage("");
  };

  const navigate = useNavigate();

  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: onOpenChannelMessage,
  });

  let response: any;

  function onOpenChannelMessage(event: MessageEvent<string>) {
    response = JSON.parse(event.data);
    console.log('response', response)
    if (response?.type === "Error") {
      response.status = "FAILED"
      setErrorMEssage(response.error);
      setResultState(ProgressStepState.error);
      return;
    }
    response.status = "SUCCEEDED"
  }

  return (
    <PopoutPageTemplate title={"Open Channel"} show={true} onClose={() => navigate(-1)} icon={<ChannelsIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Connect"} state={connectState} last={false} />
        <Step label={"Detail"} state={detailState} last={false} />
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
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
          <div className={styles.openChannelTableRow}>
            <FormRow>
              <div className={styles.openChannelTableSingle}>
                <span className={styles.label}>{"Peer public key"}</span>
                <div className={styles.input}>
                  <TextInput
                    value={nodePubKey}
                    placeholder={"pubkey"}
                    onChange={(value) => {
                      setNodePubKey(value as string);
                    }}
                  />
                </div>
              </div>
            </FormRow>
          </div>
          <div className={styles.openChannelTableRow}>
            <FormRow>
              <div className={styles.openChannelTableSingle}>
                <span className={styles.label}>{"Peer IP and port"}</span>
                <div className={styles.input}>
                  <TextInput
                    value={host}
                    placeholder={"ip:port"}
                    onChange={(value) => {
                      setHost(value as string);
                    }}
                  />
                </div>
              </div>
            </FormRow>
          </div>
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={"Comfirm"}
                disabled={ host == "" || nodePubKey == "" || selectedNodeId == 0 }
                onClick={() => {
                  setStepIndex(1);
                  setConnectState(ProgressStepState.completed);
                  setDetailState(ProgressStepState.active);
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"Channel Size"}</span>
                  <div className={styles.input}>
                    <NumberFormat
                      className={styles.single}
                      thousandSeparator={false}
                      value={localFundingAmount}
                      onValueChange={(values: NumberFormatValues) => {
                        setLocalFundingAmount(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"Sats per vbyte"}</span>
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
            <SectionContainer
            title={"Advanced Options"}
            icon={AdvencedOption}
            expanded={expandAdvancedOptions}
            handleToggle={() => {
              setExpandAdvancedOptions(!expandAdvancedOptions);
            }}
          >
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"Push Amount"}</span>
                  <div className={styles.input}>
                    <NumberFormat
                      className={styles.single}
                      thousandSeparator={false}
                      value={pushSat}
                      onValueChange={(values: NumberFormatValues) => {
                        setPushSat(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"HTLC min sat"}</span>
                  <div className={styles.input}>
                    <NumberFormat
                      className={styles.single}
                      thousandSeparator={false}
                      value={minHtlcMsat}
                      onValueChange={(values: NumberFormatValues) => {
                        setMinHtlcMsat(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"Minimum Confirmations"}</span>
                  <div className={styles.input}>
                    <NumberFormat
                      className={styles.single}
                      thousandSeparator={false}
                      value={minConfs}
                      onValueChange={(values: NumberFormatValues) => {
                        setMinConfs(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{"Channel Close Address"}</span>
                  <div className={styles.input}>
                    <TextInput
                      value={closeAddress}
                      placeholder={"e.g. bc1q..."}
                      onChange={(value) => {
                        setCloseAddress(value as string);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <div className={styles.openChannelTableRow}>
              <FormRow>
              <Switch label={"Private"}
                checked={privateChan}
                onChange={(value) => {
                  console.log('value', value)
                  setPrivate(value)
                }}
              />
              </FormRow>
              <FormRow>
              <Switch label={"Spend unconfirmed outputs"}
                checked={spendUnconfirmed}
                onChange={(value) => {
                  console.log('value', value)
                  setSpendUnconfirmed(value)
                }}
              />
              </FormRow>
            </div>
          </SectionContainer>
            <ButtonWrapper
              rightChildren={
                <Button
                  text={t.openCloseChannel.confirm}
                  onClick={() => {
                    setStepIndex(2);
                    setDetailState(ProgressStepState.completed);
                    setResultState(ProgressStepState.completed);
                    sendJsonMessage({
                      reqId: "randId",
                      type: "openChannel",
                      openChannelRequest: {
                        nodeId: selectedNodeId,
                        satPerVbyte,
                        nodePubKey,
                        host,
                        localFundingAmount,
                        pushSat,
                        private: privateChan,
                        spendUnconfirmed,
                        minHtlcMsat,
                        minConfs,
                        closeAddress,
                      },
                    })
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
              styles.openChannelResultIconWrapper,
              { [styles.failed]: errMessage },
              openStatusClass[errMessage ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {" "}
            {openStatusIcon[errMessage ? "FAILED" : "SUCCEEDED"]}
          </div>
          <div className={errMessage ? styles.errorBox : styles.successeBox }>
            <div>
              <div className={errMessage ? styles.errorIcon : styles.successIcon }>{openStatusIcon["NOTE"]}</div>
              <div className={errMessage ? styles.errorNote : styles.successNote}>{errMessage ? t.openCloseChannel.error :t.openCloseChannel.note}</div>
            </div >
            <div className={errMessage ? styles.errorMessage: styles.successMessage }>
              {errMessage ? errMessage : t.openCloseChannel.confirmationOpenning}
            </div>
          </div>
          <ButtonWrapper
            rightChildren={
              <Button
                text={t.openCloseChannel.openNewChannel}
                onClick={() => {
                  setStepIndex(0);
                  setConnectState(ProgressStepState.active);
                  setDetailState(ProgressStepState.disabled);
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

export default OpenChannelModal;
