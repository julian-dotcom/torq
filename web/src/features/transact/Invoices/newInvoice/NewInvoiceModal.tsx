import { Options20Regular as OptionsIcon, MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, useNewInvoiceMutation } from "apiSlice";
import Button, { buttonColor, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "./newInvoice.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select from "features/forms/Select";
import LargeAmountInput from "features/inputs/largeAmountInput/LargeAmountInput";
import { SectionContainer } from "features/section/SectionContainer";
import Input from "components/forms/input/Input";
import { formatDuration, intervalToDuration } from "date-fns";
import { NewInvoiceResponseStep } from "./newInvoiceResponse";

function NewInvoiceModal() {
  const { t } = useTranslations();
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const [newInvoiceMutation, newInvoiceResponse] = useNewInvoiceMutation();

  let nodeConfigurationOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations !== undefined) {
    nodeConfigurationOptions = nodeConfigurations.map((node: nodeConfiguration) => {
      return { value: node.nodeId, label: node.grpcAddress };
    });
  }

  useEffect(() => {
    if (nodeConfigurationOptions !== undefined) {
      setSelectedNodeId(nodeConfigurationOptions[0].value);
    }
    if (newInvoiceResponse.isSuccess) {
      setDoneState(ProgressStepState.completed);
    }
    if (newInvoiceResponse.isError) {
      setDoneState(ProgressStepState.error);
    }
    if (newInvoiceResponse.isLoading) {
      setDoneState(ProgressStepState.processing);
    }
  }, [nodeConfigurationOptions, newInvoiceResponse]);

  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value);
  const [amountSat, setAmountSat] = useState<number | undefined>(undefined);
  const [expirySeconds, setExpirySeconds] = useState<number | undefined>(undefined);
  const [memo, setMemo] = useState<string | undefined>(undefined);
  const [fallbackAddress, setFallbackAddress] = useState<string | undefined>(undefined);

  const [detailsState, setDetailsState] = useState(ProgressStepState.active);
  const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  const closeAndReset = () => {
    setStepIndex(0);
    setDetailsState(ProgressStepState.active);
    setDoneState(ProgressStepState.disabled);
  };

  const handleClickNext = () => {
    setStepIndex(1);
    setDetailsState(ProgressStepState.completed);
    setDoneState(ProgressStepState.processing);
    newInvoiceMutation({
      nodeId: selectedNodeId,
      valueMsat: amountSat ? amountSat * 1000 : undefined, // msat = 1000*sat
      expiry: expirySeconds,
      memo: memo,
    });
  };

  const navigate = useNavigate();

  const d = intervalToDuration({ start: 0, end: expirySeconds ? expirySeconds * 1000 : 86400 * 1000 });
  const pif = formatDuration({
    years: d.years,
    months: d.months,
    days: d.days,
    hours: d.hours,
    minutes: d.minutes,
    seconds: d.seconds,
  });

  return (
    <PopoutPageTemplate
      title={t.newInvoice.title}
      show={true}
      onClose={() => navigate(-1)}
      icon={<TransactionIconModal />}
    >
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={t.newInvoice.details} state={detailsState} last={false} />
        <Step label={t.newInvoice.invoice} state={doneState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <LargeAmountInput
            label={t.newInvoice.amount}
            value={amountSat}
            autoFocus={true}
            onChange={(value) => {
              setAmountSat(value);
            }}
          />
          <Select
            label={t.yourNode}
            onChange={(newValue: any) => {
              setSelectedNodeId(newValue?.value || 0);
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
          <textarea
            id={"destination"}
            name={"destination"}
            placeholder={"Optionally add a memo to this invoice."} // , PubKey or On-chain Address
            className={styles.textArea}
            value={memo}
            onChange={(e: ChangeEvent<HTMLTextAreaElement>) => {
              setMemo(e.target.value.toString());
            }}
            rows={6}
          />

          <SectionContainer
            title={"Advanced Options"}
            icon={OptionsIcon}
            expanded={expandAdvancedOptions}
            handleToggle={() => {
              setExpandAdvancedOptions(!expandAdvancedOptions);
            }}
          >
            <Input
              label={t.newInvoice.expiry + ` (${pif})`}
              value={expirySeconds}
              type={"text"}
              placeholder={"86,400 seconds (24 hours)"}
              onChange={(e) => {
                e ? setExpirySeconds(parseInt(e.target.value)) : setExpirySeconds(undefined);
              }}
            />
            <Input
              label={t.newInvoice.fallbackAddress}
              value={fallbackAddress}
              type={"text"}
              placeholder={"e.g. bc1q..."}
              onChange={(e) => {
                setFallbackAddress(e.target.value);
              }}
            />
          </SectionContainer>
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={"Confirm"}
                onClick={() => {
                  handleClickNext();
                }}
                buttonColor={buttonColor.green}
              />
            }
          />
        </ProgressTabContainer>

        <ProgressTabContainer>
          <NewInvoiceResponseStep
            selectedNodeId={selectedNodeId}
            amount={amountSat ? amountSat : 0}
            clearFlow={() => {
              setDetailsState(ProgressStepState.active);
              setDoneState(ProgressStepState.disabled);
              setStepIndex(0);
              newInvoiceResponse.reset();
            }}
            response={{
              data: newInvoiceResponse.data,
              error: newInvoiceResponse.error,
              isLoading: newInvoiceResponse.isLoading,
              isSuccess: newInvoiceResponse.isSuccess,
              isError: newInvoiceResponse.isError,
              isUninitialized: newInvoiceResponse.isUninitialized,
            }}
            setDoneState={setDoneState}
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewInvoiceModal;
