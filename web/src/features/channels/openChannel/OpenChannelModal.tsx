import {
  ArrowRouting20Regular as ChannelsIcon,
  ArrowSyncFilled as ProcessingIcon,
  Checkmark20Regular as SuccessNoteIcon,
  CheckmarkRegular as SuccessIcon,
  CommentLightning20Regular as AdvencedOption,
  ErrorCircleRegular as FailedIcon,
  Link20Regular as LinkIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, WS_URL } from "apiSlice";
import { ChangeEvent, useEffect, useState } from "react";
import Button, { ButtonWrapper, ColorVariant, ExternalLinkButton } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./openChannel.module.scss";
import { useNavigate } from "react-router";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import classNames from "classnames";
import { NumberFormatValues } from "react-number-format";
import Input from "components/forms/input/Input";
import { SectionContainer } from "features/section/SectionContainer";
import useWebSocket from "react-use-websocket";
import Switch from "components/forms/switch/Switch";
import { v4 as uuidv4 } from "uuid";
import FormRow from "features/forms/FormWrappers";
import Note, { NoteType } from "features/note/Note";
import { Select, TextArea } from "components/forms/forms";
import { InputSizeVariant } from "components/forms/input/variants";
import mixpanel from "mixpanel-browser";
import ErrorSummary from "components/errors/ErrorSummary";
import { FormErrors, mergeServerError } from "components/errors/errors";

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
  let nodeConfigurationOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
    });
  }

  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value as number);
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [openingTx, setOpeningTx] = useState<string>("");
  const [formErrorState, setFormErrorState] = useState({} as FormErrors);
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
  const [connectionString, setConnectionString] = useState<string>("");
  const [nodePubKey, setNodePubKey] = useState<string>("");
  const [host, setHost] = useState<string>("");
  const [stepIndex, setStepIndex] = useState(0);
  const [requestUUID, setRequestUUID] = useState("");

  useEffect(() => {
    if (nodeConfigurationOptions !== undefined) {
      setSelectedNodeId(nodeConfigurationOptions[0].value);
    }
  }, [nodeConfigurationOptions]);

  function handleNodeSelection(value: number) {
    setSelectedNodeId(value);
  }

  const closeAndReset = () => {
    setStepIndex(0);
    setSelectedNodeId(0);
    setConnectState(ProgressStepState.active);
    setDetailState(ProgressStepState.disabled);
    setResultState(ProgressStepState.disabled);
    setOpeningTx("");
    setFormErrorState({});
  };

  const navigate = useNavigate();

  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: onOpenChannelMessage,
  });

  function onOpenChannelMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    if (!response || response.id !== requestUUID) {
      return;
    }
    setStepIndex(1);
    setDetailState(ProgressStepState.completed);
    if (response?.type === "Error") {
      setFormErrorState(mergeServerError(response.error, formErrorState));
      setResultState(ProgressStepState.error);
      setStepIndex(2);
      return;
    }
    if (!openingTx) {
      setResultState(ProgressStepState.completed);
      const channelPoint: string = response.pendingChannelPoint?.substring(
        0,
        response.pendingChannelPoint?.indexOf(":")
      );
      setOpeningTx(channelPoint.trim());
    }
  }

  function openChannel() {
    setDetailState(ProgressStepState.processing);
    mixpanel.track("Open Channel", {
      nodeId: selectedNodeId,
      openChannelUseSatPerVbyte: satPerVbyte !== 0,
      openChannelUsePushAmount: pushSat !== 0,
      openChannelUseHTLCMinSat: minHtlcMsat !== 0,
      openChannelUseMinimumConfirmations: minConfs !== 0,
      openChannelUseChannelCloseAddress: closeAddress !== "",
    });
    const newRequestUUID = uuidv4();
    setRequestUUID(newRequestUUID);
    sendJsonMessage({
      requestId: newRequestUUID,
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
    });
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
          <div className={styles.openChannelTableRow}>
            <FormRow>
              <Select
                label={t.yourNode}
                onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                  const selectOptions = newValue as SelectOptions;
                  handleNodeSelection(selectOptions?.value as number);
                }}
                options={nodeConfigurationOptions}
                value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
              />
            </FormRow>
          </div>
          <div className={styles.openChannelTableRow}>
            <FormRow>
              <div className={styles.openChannelTableSingle}>
                <div className={styles.input}>
                  <TextArea
                    label={t.ConnectionString}
                    helpText={t.NodeConnectionStringHelp}
                    sizeVariant={InputSizeVariant.normal}
                    value={connectionString}
                    rows={4}
                    placeholder={
                      "03aab7e9327716ee946b8fbfae039b01235356549e72c5cca113ea67893d0821e5@123.123.123.123:9735"
                    }
                    onChange={(e) => {
                      setConnectionString(e.target.value);
                      if (e.target.value) {
                        const split = e.target.value.split("@");
                        split[0] && setNodePubKey(split[0]);
                        split[1] && setHost(split[1]);
                      }
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
                disabled={host == "" || nodePubKey == "" || selectedNodeId == 0}
                onClick={() => {
                  setStepIndex(1);
                  setConnectState(ProgressStepState.completed);
                  setDetailState(ProgressStepState.active);
                }}
                buttonColor={ColorVariant.primary}
              >
                {t.confirm}
              </Button>
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <div className={styles.openChannelTableRow}>
              <FormRow>
                <div className={styles.openChannelTableSingle}>
                  <span className={styles.label}>{t.ChannelSize}</span>
                  <div className={styles.input}>
                    <Input
                      formatted={true}
                      className={styles.single}
                      thousandSeparator={","}
                      value={localFundingAmount}
                      suffix={" sat"}
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
                  <div className={styles.input}>
                    <Input
                      label={t.SatPerVbyte}
                      formatted={true}
                      className={styles.single}
                      thousandSeparator={","}
                      value={satPerVbyte}
                      suffix={" sat"}
                      onValueChange={(values: NumberFormatValues) => {
                        setSatPerVbyte(values.floatValue as number);
                      }}
                    />
                  </div>
                </div>
              </FormRow>
            </div>
            <SectionContainer
              title={t.AdvancedOptions}
              icon={AdvencedOption}
              expanded={expandAdvancedOptions}
              handleToggle={() => {
                setExpandAdvancedOptions(!expandAdvancedOptions);
              }}
            >
              <div className={styles.openChannelTableRow}>
                <FormRow>
                  <div className={styles.openChannelTableSingle}>
                    <div className={styles.input}>
                      <Input
                        label={t.PushAmount}
                        formatted={true}
                        className={styles.single}
                        helpText={t.PushAmountHelpText}
                        thousandSeparator={","}
                        suffix={" sat"}
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
                    <div className={styles.input}>
                      <Input
                        label={t.HTLCMinSat}
                        formatted={true}
                        className={styles.single}
                        suffix={" sat"}
                        thousandSeparator={","}
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
                    <div className={styles.input}>
                      <Input
                        label={t.MinimumConfirmations}
                        formatted={true}
                        className={styles.single}
                        thousandSeparator={","}
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
                    <div className={styles.input}>
                      <Input
                        label={t.ChannelCloseAddress}
                        value={closeAddress}
                        type={"text"}
                        placeholder={"e.g. bc1q..."}
                        onChange={(e: ChangeEvent<HTMLInputElement>) => {
                          setCloseAddress(e.target.value);
                        }}
                      />
                    </div>
                  </div>
                </FormRow>
              </div>
              <div className={styles.openChannelTableRow}>
                <FormRow className={styles.switchRow}>
                  <Switch
                    label={t.Private}
                    checked={privateChan}
                    onChange={(e: ChangeEvent<HTMLInputElement>) => {
                      setPrivate(e.target.checked);
                    }}
                  />
                </FormRow>
                <FormRow className={styles.switchRow}>
                  <Switch
                    label={"Spend unconfirmed outputs"}
                    checked={spendUnconfirmed}
                    onChange={(e: ChangeEvent<HTMLInputElement>) => {
                      setSpendUnconfirmed(e.target.checked);
                    }}
                  />
                </FormRow>
              </div>
            </SectionContainer>
            <ButtonWrapper
              rightChildren={
                <Button onClick={openChannel} buttonColor={ColorVariant.success}>
                  {t.confirm}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.openChannelResultIconWrapper,
              { [styles.failed]: resultState === ProgressStepState.completed },
              openStatusClass[resultState === ProgressStepState.completed ? "SUCCEEDED" : "FAILED"]
            )}
          >
            {openStatusIcon[resultState === ProgressStepState.completed ? "SUCCEEDED" : "FAILED"]}
          </div>
          <div className={styles.closeChannelResultDetails}>
            {resultState === ProgressStepState.completed && (
              <>
                <Note title={t.TxId} icon={<SuccessNoteIcon />} noteType={NoteType.success}>
                  {openingTx}
                </Note>
                <ExternalLinkButton
                  href={"https://mempool.space/tx/" + openingTx}
                  target="_blank"
                  rel="noreferrer"
                  buttonColor={ColorVariant.success}
                  icon={<LinkIcon />}
                >
                  {t.openCloseChannel.GoToMempool}
                </ExternalLinkButton>

                <Note title={t.note} icon={<NoteIcon />} noteType={NoteType.info}>
                  {t.openCloseChannel.confirmationOpenning}
                </Note>
              </>
            )}

            <ErrorSummary errors={formErrorState} />

            <ButtonWrapper
              rightChildren={
                <Button
                  onClick={() => {
                    closeAndReset();
                  }}
                  buttonColor={ColorVariant.primary}
                >
                  {t.openCloseChannel.openNewChannel}
                </Button>
              }
            />
          </div>
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default OpenChannelModal;
