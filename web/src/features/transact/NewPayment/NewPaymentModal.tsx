import { MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, WS_URL } from "apiSlice";
import { useGetDecodedInvoiceQuery, useSendOnChainMutation } from "./newPaymentApi";
import classNames from "classnames";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import { NewPaymentError, NewPaymentResponse } from "features/transact/Payments/paymentTypes";
import styles from "./newPayments.module.scss";
import { PaymentProcessingErrors } from "./paymentErrorMessages";
import OnChanPaymentDetails from "./OnChanPaymentDetails";
import { PaymentType, PaymentTypeLabel } from "./types";
import { nodeConfiguration } from "apiTypes";
import Select from "features/forms/Select";
import useTranslations from "services/i18n/useTranslations";
import InvoicePayment from "./InvoicePayment";
import { InvoicePaymentResponse } from "./InvoicePaymentResponse";
import { OnChainPaymentResponse } from "./OnChainPaymentResponse";
import Note from "features/note/Note";

// RegEx used to check what type of destination the user enters.
// You can test them out here: https://regex101.com/r/OiXAlz/1
const LnPayrequestMainnetRegEx = /lnbc[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestTestnetRegEx = /lntb[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestSignetRegEx = /lnsb[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestRegtestRegEx = /lnbcrt[0-9][0-9a-zA-Z]*/gm;
const P2SHAddressRegEx = /^[3/r][0-9a-zA-Z]*/gm; // Pay to Script Hash
const P2WKHAddressRegEx = /^bc1q[0-9a-zA-Z]*/gm; // Segwit address
const P2TRAddressRegEx = /^bc1p[0-9a-zA-Z]*/gm; // Taproot address

const P2WKHAddressSignetRegEx = /^sb1q[0-9a-zA-Z]*/gm; // Segwit address
const P2TRAddressSignetRegEx = /^sb1p[0-9a-zA-Z]*/gm; // Taproot address
// const LightningNodePubkeyRegEx = /^[0-9a-fA-F]{66}$/gm; // Keysend / Lightning Node Pubkey

function NewPaymentModal() {
  const { t } = useTranslations();
  const [lnInvoiceResponses, setLnInvoiceResponses] = useState<Array<NewPaymentResponse>>([]);
  const [sendCoinsMutation, response] = useSendOnChainMutation();

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  let nodeConfigurationOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations !== undefined) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
    });
  }

  const [selectedNodeId, setSelectedNodeId] = useState<number>(0);
  const [destination, setDestination] = useState("");
  const [destinationType, setDestinationType] = useState<PaymentType>(0);
  const [overriddenLnAmount, setOverriddenLnAmount] = useState(0);
  const [destState, setDestState] = useState(ProgressStepState.active);
  const [confirmState, setConfirmState] = useState(ProgressStepState.disabled);
  const [processState, setProcessState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  const [paymentProcessingError, setPaymentProcessingError] = useState("");

  // const [paymentDescription, setPaymentDescription] = useState("");
  const [amount, setAmount] = useState<number>(0);

  useEffect(() => {
    if (nodeConfigurationOptions !== undefined) {
      setSelectedNodeId(nodeConfigurationOptions[0].value);
    }
  }, [nodeConfigurationOptions]);

  function onNewPaymentMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    if (response?.type === "Error") {
      setPaymentProcessingError(response.error);
      onNewPaymentError(response as NewPaymentError);
      return;
    }
    onNewPaymentResponse(response as NewPaymentResponse);
  }

  function onNewPaymentResponse(message: NewPaymentResponse) {
    setLnInvoiceResponses((prev) => [...prev, message]);
    if (message.status === "SUCCEEDED") {
      setProcessState(ProgressStepState.completed);
    } else if (message.status === "FAILED") {
      setPaymentProcessingError(message.failureReason);
      setProcessState(ProgressStepState.error);
    }
  }

  function onNewPaymentError(message: NewPaymentError) {
    console.log(PaymentProcessingErrors.get(message.error), message.error);
    setProcessState(ProgressStepState.error);
  }

  // This can also be an async getter function. See notes below on Async Urls.
  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: onNewPaymentMessage,
  });

  // const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  // On Chain advanced options
  const isLnInvoice = [
    PaymentType.LightningMainnet,
    PaymentType.LightningRegtest,
    PaymentType.LightningSimnet,
    PaymentType.LightningTestnet,
  ].includes(destinationType);
  const isOnChain = [PaymentType.P2SH, PaymentType.P2TR, PaymentType.P2WKH, PaymentType.P2PKH].includes(
    destinationType
  );

  // TODO: Get the estimated fee as well
  const decodedInvRes = useGetDecodedInvoiceQuery(
    { invoice: destination, nodeId: selectedNodeId },
    {
      skip: !isLnInvoice,
    }
  );

  const closeAndReset = () => {
    setStepIndex(0);
    setDestState(ProgressStepState.active);
    setConfirmState(ProgressStepState.disabled);
    setProcessState(ProgressStepState.disabled);
  };

  const setDestinationHandler = (e: ChangeEvent<HTMLTextAreaElement>) => {
    if (e.target.value !== destination) {
      setDestinationType(0);
      decodedInvRes.data = undefined;
      // setPaymentDescription("");
    }

    setDestination(e.target.value);
    if (e.target.value === "") {
      setDestinationType(0);
      setDestState(ProgressStepState.active);
      return;
    } else if (e.target.value.match(LnPayrequestMainnetRegEx)) {
      setDestinationType(PaymentType.LightningMainnet);
    } else if (e.target.value.match(LnPayrequestTestnetRegEx)) {
      setDestinationType(PaymentType.LightningTestnet);
    } else if (e.target.value.match(LnPayrequestSignetRegEx)) {
      setDestinationType(PaymentType.LightningSimnet);
    } else if (e.target.value.match(LnPayrequestRegtestRegEx)) {
      setDestinationType(PaymentType.LightningRegtest);
    } else if (e.target.value.match(P2SHAddressRegEx)) {
      // Pay to Script Hash
      setDestinationType(PaymentType.P2SH);
    } else if (e.target.value.match(P2WKHAddressRegEx) || e.target.value.match(P2WKHAddressSignetRegEx)) {
      // Segwit address
      setDestinationType(PaymentType.P2WKH);
    } else if (e.target.value.match(P2TRAddressRegEx) || e.target.value.match(P2TRAddressSignetRegEx)) {
      // Taproot
      setDestinationType(PaymentType.P2TR);
      // } else if (e.target.value.match(LightningNodePubkeyRegEx)) { // TODO: Add support for Keysend
      //   setDestinationType(PaymentType.Keysend);
    } else {
      setDestinationType(PaymentType.Unknown);
    }

    // Prevent accidentally adding additional characters to the destination field after
    // the user has entered a valid destination by unfocusing (bluring) the input field.
    setStepIndex(1);
    setDestState(ProgressStepState.completed);
    setConfirmState(ProgressStepState.active);
  };

  const clearPaymentFlow = () => {
    setLnInvoiceResponses([]);
    setDestinationType(0);
    setDestination("");
    setDestState(ProgressStepState.active);
    setConfirmState(ProgressStepState.disabled);
    setProcessState(ProgressStepState.disabled);
    setPaymentProcessingError("");
    setStepIndex(0);
  };

  const dynamicConfirmedState = () => {
    if (decodedInvRes.isLoading || decodedInvRes.isFetching) {
      return ProgressStepState.disabled;
    } else if (decodedInvRes.isError) {
      return ProgressStepState.disabled;
    }
    return confirmState;
  };

  const dynamicDestinationState = () => {
    if (decodedInvRes.isLoading || decodedInvRes.isFetching) {
      return ProgressStepState.processing;
    } else if (decodedInvRes.isError) {
      return ProgressStepState.error;
    }
    return destState;
  };

  const navigate = useNavigate();

  function destiationLabel() {
    if (decodedInvRes?.isError === false) {
      return PaymentTypeLabel[destinationType] + " Detected";
    }
    return "Can't decode invoice";
  }

  const lnAmount = overriddenLnAmount !== 0 ? overriddenLnAmount : (decodedInvRes?.data?.valueMsat ?? 0) / 1000;

  return (
    <PopoutPageTemplate
      title={t.header.newPayment}
      show={true}
      onClose={() => navigate(-1)}
      icon={<TransactionIconModal />}
    >
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Destination"} state={dynamicDestinationState()} last={false} />
        <Step label={"Details"} state={dynamicConfirmedState()} last={false} />
        <Step label={"Process"} state={processState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <Select
            label={t.yourNode}
            onChange={
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              (newValue: any) => {
                setSelectedNodeId(newValue?.value);
              }
            }
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
          <div className={styles.destination}>
            <div className={styles.destinationWrapper}>
              <div className={styles.labelWrapper}>
                <label htmlFor={"destination"} className={styles.destinationLabel}>
                  Destination
                </label>
                {destination && (
                  <span
                    className={classNames(styles.destinationType, {
                      [styles.unknownAddressType]: destinationType === PaymentType.Unknown,
                      [styles.errorDestination]: decodedInvRes.isError,
                    })}
                  >
                    {destiationLabel()}
                  </span>
                )}
              </div>
              <textarea
                id={"destination"}
                name={"destination"}
                disabled={!selectedNodeId}
                placeholder={"E.g. Lightning Invoice or on-chain address"} // , PubKey or On-chain Address
                className={styles.destinationTextArea}
                value={destination}
                onChange={setDestinationHandler}
                rows={6}
              />
            </div>
          </div>
          <Note title={"Note:"}>
            <span>
              Torq will detect the transaction type based on the destination you enter. Valid destinations are on-chain
              addresses and lightning invoices.
            </span>
          </Note>
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                disabled={!destinationType || decodedInvRes.isError}
                onClick={() => {
                  if (destination) {
                    setStepIndex(1);
                    setDestState(ProgressStepState.completed);
                    setConfirmState(ProgressStepState.active);
                  }
                }}
                buttonColor={decodedInvRes?.isError ? ColorVariant.warning : ColorVariant.primary}
              >
                {t.next}
              </Button>
            }
          />
        </ProgressTabContainer>
        {isLnInvoice && decodedInvRes.data && (
          <InvoicePayment
            decodedInvoice={decodedInvRes.data}
            destination={destination}
            destinationType={destinationType}
            setConfirmState={setConfirmState}
            setProcessState={setProcessState}
            setStepIndex={setStepIndex}
            setDestState={setDestState}
            selectedNodeId={selectedNodeId}
            sendJsonMessage={sendJsonMessage}
            onAmountChange={setOverriddenLnAmount}
          />
        )}
        {isOnChain && (
          <OnChanPaymentDetails
            amount={amount}
            setAmount={setAmount}
            sendCoinsMutation={sendCoinsMutation}
            destination={destination}
            destinationType={destinationType}
            setConfirmState={setConfirmState}
            setStepIndex={setStepIndex}
            setProcessState={setProcessState}
            setDestState={setDestState}
          />
        )}
        {isLnInvoice && decodedInvRes.data && (
          <InvoicePaymentResponse
            selectedNodeId={selectedNodeId}
            paymentProcessingError={PaymentProcessingErrors.get(paymentProcessingError) || paymentProcessingError}
            decodedInvoice={decodedInvRes.data}
            destination={destination}
            responses={lnInvoiceResponses}
            clearPaymentFlow={clearPaymentFlow}
            amount={lnAmount}
          />
        )}
        {isOnChain && response.data && (
          <OnChainPaymentResponse
            amount={amount}
            selectedNodeId={selectedNodeId}
            setProcessState={setProcessState}
            response={{
              data: response.data,
              isUninitialized: response.isUninitialized,
              isLoading: response.isLoading,
              isSuccess: response.isSuccess,
              error: response.error,
            }}
            destination={destination}
            clearPaymentFlow={clearPaymentFlow}
          />
        )}
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewPaymentModal;
