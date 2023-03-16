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
import { useGetNodeConfigurationsQuery } from "apiSlice";
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
import Switch from "components/forms/switch/Switch";
import FormRow from "features/forms/FormWrappers";
import Note, { NoteType } from "features/note/Note";
import { Select, TextArea } from "components/forms/forms";
import { InputSizeVariant } from "components/forms/input/variants";
import mixpanel from "mixpanel-browser";
import { useOpenChannelMutation } from "./openChannelApi";
import { RtqToServerError } from "components/errors/errors";
import ErrorSummary from "components/errors/ErrorSummary";

const openStatusClass = {
  PROCESSING: styles.processing,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const openStatusIcon = {
  PROCESSING: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};

function isOption(result: unknown): result is SelectOptions & { value: number } {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as SelectOptions).value === "number"
  );
}

function OpenChannelModal() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const [resultState, setResultState] = useState(ProgressStepState.disabled);
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);
  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<Array<SelectOptions>>();
  const [connectState, setConnectState] = useState(ProgressStepState.active);
  const [stepIndex, setStepIndex] = useState(0);

  const [selectedNodeId, setSelectedNodeId] = useState<number | undefined>();
  const [detailState, setDetailState] = useState(ProgressStepState.disabled);
  const [minConfs, setMinConfs] = useState<number | undefined>();
  const [localFundingAmount, setLocalFundingAmount] = useState<number>(0);
  const [pushSat, setPushSat] = useState<number | undefined>();
  const [minHtlcMsat, setMinHtlcMsat] = useState<number | undefined>();
  const [closeAddress, setCloseAddress] = useState<string | undefined>();
  const [spendUnconfirmed, setSpendUnconfirmed] = useState<boolean>(false);
  const [privateChan, setPrivate] = useState<boolean>(false);
  const [satPerVbyte, setSatPerVbyte] = useState<number | undefined>();
  const [connectionString, setConnectionString] = useState<string | undefined>();
  const [nodePubKey, setNodePubKey] = useState<string>("");
  const [host, setHost] = useState<string | undefined>();

  const [openChannel, { data: openChannelResponse, error: openChannelError, isError, isLoading, isSuccess }] =
    useOpenChannelMutation();

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  useEffect(() => {
    if (nodeConfigurations) {
      const options = nodeConfigurations.map((node: nodeConfiguration) => {
        return { label: node.name, value: node.nodeId };
      });
      setNodeConfigurationOptions(options);
      setSelectedNodeId(options[0].value);
    }
  }, [nodeConfigurations]);

  useEffect(() => {
    if (isSuccess) {
      setResultState(ProgressStepState.completed);
    }
    if (isLoading) {
      setResultState(ProgressStepState.processing);
    }
    if (isError) {
      setResultState(ProgressStepState.error);
    }
  }, [isSuccess, isError, isLoading]);

  function handleNodeSelection(value: number) {
    setSelectedNodeId(value);
  }

  const closeAndReset = () => {
    setStepIndex(0);
    setConnectState(ProgressStepState.active);
    setDetailState(ProgressStepState.disabled);
    setResultState(ProgressStepState.disabled);

    setNodePubKey("");
    setLocalFundingAmount(0);

    setExpandAdvancedOptions(false);
    setMinConfs(undefined);
    setPushSat(undefined);
    setMinHtlcMsat(undefined);
    setCloseAddress(undefined);
    setSpendUnconfirmed(false);
    setPrivate(false);
    setSatPerVbyte(undefined);
    setConnectionString("");
    setHost(undefined);
  };

  function handleOpenChannel() {
    if (!selectedNodeId) return;

    setStepIndex(2);
    setDetailState(ProgressStepState.completed);
    setResultState(ProgressStepState.processing);
    mixpanel.track("Open Channel", {
      nodeId: selectedNodeId,
      openChannelUseSatPerVbyte: satPerVbyte !== 0,
      openChannelUsePushAmount: pushSat !== 0,
      openChannelUseHTLCMinSat: minHtlcMsat !== 0,
      openChannelUseMinimumConfirmations: minConfs !== 0,
      openChannelUseChannelCloseAddress: closeAddress !== "",
    });
    openChannel({
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
                  // Check if newValue is of type SelectOptions
                  if (isOption(newValue)) {
                    const selectOptions = newValue as SelectOptions;
                    handleNodeSelection(selectOptions?.value as number);
                  }
                }}
                options={nodeConfigurationOptions}
                value={nodeConfigurationOptions?.find((option) => option.value === selectedNodeId)}
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
                    placeholder={"03aab7e9327716ee946b8fbfae039b01235356549e72c5cca113ea67893d0821e5@123.1.3.65:9735"}
                    onChange={(e) => {
                      setConnectionString(e.target.value);
                      if (!e.target.value) {
                        setNodePubKey("");
                      }
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
                disabled={host == "" || nodePubKey == "" || selectedNodeId === undefined}
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
                <Button onClick={handleOpenChannel} buttonColor={ColorVariant.success}>
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
              { [styles.failed]: isError },
              openStatusClass[isLoading ? "PROCESSING" : isError ? "FAILED" : "SUCCEEDED"]
            )}
          >
            {openStatusIcon[isLoading ? "PROCESSING" : isError ? "FAILED" : "SUCCEEDED"]}
          </div>
          <div className={styles.closeChannelResultDetails}>
            {isLoading && (
              <Note title={t.Processing} icon={<ProcessingIcon />} noteType={NoteType.warning}>
                {t.openCloseChannel.processingClose}
              </Note>
            )}
            {isSuccess && (
              <>
                <Note title={t.TxId} icon={<SuccessNoteIcon />} noteType={NoteType.success}>
                  {openChannelResponse?.fundingTransactionHash}
                </Note>
                <ExternalLinkButton
                  href={"https://mempool.space/tx/" + openChannelResponse?.fundingTransactionHash}
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
            {isError && <ErrorSummary title={t.Error} errors={RtqToServerError(openChannelError).errors} />}
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
