import { Options20Regular as OptionsIcon, MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetLocalNodesQuery, useNewInvoiceMutation } from "apiSlice";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "./newInvoice.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { localNode } from "apiTypes";
import Select from "features/forms/Select";
import LargeAmountInput from "features/inputs/largeAmountInput/LargeAmountInput";
import { SectionContainer } from "features/section/SectionContainer";
import TextInput from "features/forms/TextInput";
import { formatDuration, intervalToDuration } from "date-fns";
import { NewInvoiceResponseStep } from "./newInvoiceResponse";

function NewInvoiceModal() {
  const { t } = useTranslations();
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  const { data: localNodes } = useGetLocalNodesQuery();
  const [newInvoiceMutation, newInvoiceResponse] = useNewInvoiceMutation();

  let localNodeOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (localNodes !== undefined) {
    localNodeOptions = localNodes.map((localNode: localNode) => {
      return { value: localNode.localNodeId, label: localNode.grpcAddress };
    });
  }

  useEffect(() => {
    if (localNodeOptions !== undefined) {
      setSelectedLocalNode(localNodeOptions[0].value);
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
  }, [localNodeOptions, newInvoiceResponse]);

  const [selectedLocalNode, setSelectedLocalNode] = useState<number>(localNodeOptions[0].value);
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
      localNodeId: selectedLocalNode,
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
              setSelectedLocalNode(newValue?.value || 0);
            }}
            options={localNodeOptions}
            value={localNodeOptions.find((option) => option.value === selectedLocalNode)}
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
            <TextInput
              label={t.newInvoice.expiry + ` (${pif})`}
              value={expirySeconds}
              placeholder={"86,400 seconds (24 hours)"}
              onChange={(value) => {
                value ? setExpirySeconds(parseInt(value as string)) : setExpirySeconds(undefined);
              }}
            />
            <TextInput
              label={t.newInvoice.fallbackAddress}
              value={fallbackAddress}
              placeholder={"e.g. bc1q..."}
              onChange={(value) => {
                setFallbackAddress(value as string);
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
            selectedLocalNode={selectedLocalNode}
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
