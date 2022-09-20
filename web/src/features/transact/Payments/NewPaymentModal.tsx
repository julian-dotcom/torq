import styles from "./newPayments.module.scss";
import {
  MoneyHand24Regular as TransactionIconModal,
  Options20Regular as OptionsIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  ArrowSyncFilled as ProcessingIcon,
} from "@fluentui/react-icons";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import TextInput from "features/forms/TextInput";
import { SectionContainer } from "../../section/SectionContainer";
import { ChangeEvent, useState } from "react";
import Switch from "../../inputs/Slider/Switch";
import PopoutPageTemplate from "../../templates/popoutPageTemplate/PopoutPageTemplate";
import ProgressHeader, { ProgressStepState, Step } from "../../progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "../../progressTabs/ProgressTab";
import { useGetDecodedInvoiceQuery } from "apiSlice";
import { format } from "d3";
import Note from "../../note/Note";
import classNames from "classnames";
import NumberFormat, { NumberFormatValues } from "react-number-format";
import useWebSocket from "react-use-websocket";
import { WS_URL } from "apiSlice";
import { NewPaymentResponse, NewPaymentError } from "./paymentTypes";

const fd = format(",.0f");

export type NewPaymentRequest = {
  invoice: string;
  timeOutSecs: number;
  dest: string;
  amtMSat: number;
  feeLimitMsat: number;
  allowSelfPayment: boolean;
};

type NewPaymentModalProps = {
  show: boolean;
  modalCloseHandler: () => void;
};

enum PaymentType {
  Unknown,
  Keysend,
  P2PKH,
  P2SH,
  P2WKH,
  P2TR,
  LightningMainnet,
  LightningTestnet,
  LightningSimnet,
  LightningRegtest,
}

const PaymentTypeLabel = {
  [PaymentType.Unknown]: "Unknown ",
  [PaymentType.P2PKH]: "Legacy Bitcoin ", // Legacy address
  [PaymentType.P2SH]: "Pay-to-Script-Hash ", // P2SH address
  [PaymentType.P2WKH]: "Segwit ", // Segwit address
  [PaymentType.P2TR]: "Taproot Address", // Taproot address
  [PaymentType.LightningMainnet]: "Mainnet Invoice",
  [PaymentType.LightningTestnet]: "Testnet Invoice",
  [PaymentType.LightningSimnet]: "Simnet Invoice",
  [PaymentType.LightningRegtest]: "Regtest Invoice",
  [PaymentType.Keysend]: "Keysend",
};

const paymentStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const paymentStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
};

// RegEx used to check what type of destination the user enters.
// You can test them out here: https://regex101.com/r/OiXAlz/1
const LnPayrequestMainnetRegEx = /lnbc[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestTestnetRegEx = /lntb[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestSignetRegEx = /lnsb[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestRegtestRegEx = /lnbcrt[0-9][0-9a-zA-Z]*/gm;
const P2PKHAddressRegEx = /^1[0-9a-zA-Z]*/gm; // Legacy Adresses
const P2SHAddressRegEx = /^[3/r][0-9a-zA-Z]*/gm; // Pay to Script Hash
const P2WKHAddressRegEx = /^bc1q[0-9a-zA-Z]*/gm; // Segwit address
const P2TRAddressRegEx = /^bc1p[0-9a-zA-Z]*/gm; // Taproot address

const P2WKHAddressSignetRegEx = /^sb1q[0-9a-zA-Z]*/gm; // Segwit address
const P2TRAddressSignetRegEx = /^sb1p[0-9a-zA-Z]*/gm; // Taproot address

const LightningNodePubkeyRegEx = /^[0-9a-fA-F]{66}$/gm; // Keysend / Lightning Node Pubkey

function NewPaymentModal(props: NewPaymentModalProps) {
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);
  const [responses, setResponses] = useState<Array<NewPaymentResponse>>([]);
  const [destinationType, setDestinationType] = useState<PaymentType>(0);
  const [destination, setDestination] = useState("");
  const [destState, setDestState] = useState(ProgressStepState.active);
  const [confirmState, setConfirmState] = useState(ProgressStepState.disabled);
  const [processState, setProcessState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  const [paymentDescription, setPaymentDescription] = useState("");

  const [onChainPaymentAmount, setOnChainPaymentAmount] = useState<number>();

  // This can also be an async getter function. See notes below on Async Urls.
  const { sendMessage, sendJsonMessage, lastMessage, lastJsonMessage, readyState, getWebSocket } = useWebSocket(
    WS_URL,
    {
      onOpen: () => console.log("opened"),
      //Will attempt to reconnect on all close events, such as server shutting down
      shouldReconnect: (closeEvent) => true,
      share: true,
      onMessage: (event) => {
        const respnse = JSON.parse(event.data) as NewPaymentResponse | NewPaymentError;
        if (respnse?.type == "Error") {
          console.log("error", event.data);
          setProcessState(ProgressStepState.error);
          return;
        } else if (respnse?.type == "newPayment") {
          setResponses((prev) => [...prev, respnse]);
          if (respnse.status == "SUCCEEDED") {
            setProcessState(ProgressStepState.completed);
          } else if (respnse.status == "FAILED") {
            setProcessState(ProgressStepState.error);
          }
        }
      },
    }
  );

  const handleAdvancedToggle = () => {
    setExpandAdvancedOptions(!expandAdvancedOptions);
  };

  const isLnInvoice = [
    PaymentType.LightningMainnet,
    PaymentType.LightningRegtest,
    PaymentType.LightningSimnet,
    PaymentType.LightningTestnet,
  ].includes(destinationType);

  // TODO: Get the estimated fee as well
  const decodedInvRes = useGetDecodedInvoiceQuery(
    { invoice: destination },
    {
      skip: !isLnInvoice,
    }
  );

  const closeAndReset = () => {
    setStepIndex(0);
    setDestState(ProgressStepState.active);
    setConfirmState(ProgressStepState.disabled);
    setProcessState(ProgressStepState.disabled);
    props.modalCloseHandler();
  };

  const setDestinationHandler = (e: ChangeEvent<HTMLTextAreaElement>) => {
    if (e.target.value !== destination) {
      setDestinationType(0);
      decodedInvRes.data = undefined;
      setPaymentDescription("");
    }

    setDestination(e.target.value);
    if (e.target.value === "") {
      setDestinationType(0);
      return;
    } else if (e.target.value.match(LnPayrequestMainnetRegEx)) {
      setDestinationType(PaymentType.LightningMainnet);
    } else if (e.target.value.match(LnPayrequestTestnetRegEx)) {
      setDestinationType(PaymentType.LightningTestnet);
    } else if (e.target.value.match(LnPayrequestSignetRegEx)) {
      setDestinationType(PaymentType.LightningSimnet);
    } else if (e.target.value.match(LnPayrequestRegtestRegEx)) {
      setDestinationType(PaymentType.LightningRegtest);
    } else if (e.target.value.match(P2PKHAddressRegEx)) {
      setDestinationType(PaymentType.P2PKH);
    } else if (e.target.value.match(P2SHAddressRegEx)) {
      // Pay to Script Hash
      setDestinationType(PaymentType.P2SH);
    } else if (e.target.value.match(P2WKHAddressRegEx) || e.target.value.match(P2WKHAddressSignetRegEx)) {
      // Segwit address
      setDestinationType(PaymentType.P2WKH);
    } else if (e.target.value.match(P2TRAddressRegEx) || e.target.value.match(P2TRAddressSignetRegEx)) {
      // Taproot
      setDestinationType(PaymentType.P2TR);
    } else if (e.target.value.match(LightningNodePubkeyRegEx)) {
      setDestinationType(PaymentType.Keysend);
    } else {
      setDestinationType(PaymentType.Unknown);
      return;
    }

    // Prevent accidentally adding additional characters to the destination field after
    // the user has entered a valid destination by unfocusing (bluring) the input field.
    e.target.blur();
    setStepIndex(1);
    setConfirmState(ProgressStepState.active);
    setDestState(ProgressStepState.completed);
  };

  const dynamicConfirmedState = () => {
    if (decodedInvRes.isLoading || decodedInvRes.isFetching) {
      return ProgressStepState.processing;
    } else if (decodedInvRes.isError) {
      return ProgressStepState.error;
    }
    return confirmState;
  };

  const lnStep = (
    <ProgressTabContainer>
      <div className={styles.amountWrapper}>
        {/*<div className={styles.label}>You are paying</div>*/}
        {destinationType && (
          <span className={styles.destinationType}>{PaymentTypeLabel[destinationType] + " Detected"}</span>
        )}
        <div className={styles.amount}>
          {fd(
            !decodedInvRes.isLoading && decodedInvRes.data && !decodedInvRes.error
              ? decodedInvRes.data?.valueMsat / 1000
              : 0
          ) + " sat"}
        </div>
        <div className={styles.label}>To</div>
        <div className={styles.destinationPreview}>{decodedInvRes?.data?.nodeAlias}</div>
      </div>
      <div className={styles.destinationWrapper}>
        <div className={styles.labelWrapper}>
          <label htmlFor={"destination"} className={styles.destinationLabel}>
            Description (only seen by you)
          </label>
        </div>
        <textarea
          id={"lnDescription"}
          name={"lnDescription"}
          className={styles.destinationTextArea}
          autoComplete="off"
          value={paymentDescription}
          onChange={(e) => {
            setPaymentDescription(e.target.value);
          }}
          rows={3}
        />
      </div>
      <SectionContainer
        title={"Advanced Options"}
        icon={OptionsIcon}
        expanded={expandAdvancedOptions}
        handleToggle={handleAdvancedToggle}
      >
        <Switch label={"Allow self payment"} labelPosition={"left"} checked={true} />
        <TextInput label={"Max fee"} />
        <TextInput label={"Max fee rate"} />
        <TextInput label={"Timeout"} />
      </SectionContainer>

      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        leftChildren={
          <Button
            text={"Back"}
            onClick={() => {
              setStepIndex(0);
              setDestState(ProgressStepState.completed);
              setConfirmState(ProgressStepState.active);
            }}
            buttonColor={buttonColor.ghost}
          />
        }
        rightChildren={
          <Button
            text={"Confirm"}
            onClick={() => {
              setStepIndex(2);
              setConfirmState(ProgressStepState.completed);
              setProcessState(ProgressStepState.processing);
              sendJsonMessage({
                reqId: "randId",
                type: "newPayment",
                NewPaymentRequest: { invoice: destination, timeOutSecs: 3600 },
              });
            }}
            buttonColor={buttonColor.green}
          />
        }
      />
    </ProgressTabContainer>
  );

  const btcStep = (
    <ProgressTabContainer>
      <div className={styles.amountWrapper}>
        {/*<div className={styles.label}>You are paying</div>*/}
        {destinationType && (
          <span className={styles.destinationType}>{PaymentTypeLabel[destinationType] + " Detected"}</span>
        )}
        <div className={styles.amount}>
          {/*{fd(decodedInvRes.data ? decodedInvRes.data?.valueMsat / 1000 : 0) + " sat"}*/}
          <NumberFormat
            className={styles.amountInput}
            suffix={" sat"}
            thousandSeparator=","
            value={onChainPaymentAmount}
            placeholder={"0 sat"}
            onValueChange={(values: NumberFormatValues) => {
              console.log(values);
              setOnChainPaymentAmount(values.floatValue || 0);
            }}
          />
        </div>
        <div className={styles.label}>To</div>
        <div className={styles.destinationPreview}>{destination}</div>
      </div>
      <div className={styles.destinationWrapper}>
        <div className={styles.labelWrapper}>
          <label htmlFor={"destination"} className={styles.destinationLabel}>
            Description (only seen by you)
          </label>
        </div>
        <textarea
          id={"lnDescription"}
          name={"lnDescription"}
          className={styles.destinationTextArea}
          autoComplete="off"
          value={paymentDescription}
          onChange={(e) => {
            setPaymentDescription(e.target.value);
          }}
          rows={3}
        />
      </div>
      <SectionContainer
        title={"Advanced Options"}
        icon={OptionsIcon}
        expanded={expandAdvancedOptions}
        handleToggle={handleAdvancedToggle}
      >
        <TextInput label={"Sat per vByte"} />
      </SectionContainer>

      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        leftChildren={
          <Button
            text={"Back"}
            onClick={() => {
              setStepIndex(0);
              setDestState(ProgressStepState.completed);
              setConfirmState(ProgressStepState.active);
            }}
            buttonColor={buttonColor.ghost}
          />
        }
        rightChildren={
          <Button
            text={"Confirm"}
            onClick={() => {
              setStepIndex(2);
              setConfirmState(ProgressStepState.completed);
              setProcessState(ProgressStepState.processing);
            }}
            buttonColor={buttonColor.green}
          />
        }
      />
    </ProgressTabContainer>
  );

  return (
    <PopoutPageTemplate title={"New Payment"} show={props.show} onClose={closeAndReset} icon={<TransactionIconModal />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Destination"} state={destState} last={false} />
        <Step label={"Confirm"} state={dynamicConfirmedState()} last={false} />
        <Step label={"Paying"} state={processState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
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
                    })}
                  >
                    {PaymentTypeLabel[destinationType] + " Detected"}
                  </span>
                )}
              </div>
              <textarea
                id={"destination"}
                name={"destination"}
                placeholder={"E.g. Lightning Invoice, PubKey or On-chain Address"}
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
              addresses, lightning invoices, and lightning node public keys for keysend.
            </span>
          </Note>
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={"Next"}
                disabled={!destinationType}
                onClick={() => {
                  if (destination) {
                    setStepIndex(1);
                    setDestState(ProgressStepState.completed);
                    setConfirmState(ProgressStepState.active);
                  }
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
        {isLnInvoice && lnStep}
        {!isLnInvoice && btcStep}
        <ProgressTabContainer>
          <div
            className={classNames(
              styles.paymentResultIconWrapper,
              { [styles.failed]: responses.length === 0 },
              paymentStatusClass[responses[responses.length - 1]?.status as "SUCCEEDED" | "FAILED" | "IN_FLIGHT"]
            )}
          >
            {" "}
            {responses.length === 0 && paymentStatusIcon["FAILED"]}
            {paymentStatusIcon[responses[responses.length - 1]?.status as "SUCCEEDED" | "FAILED" | "IN_FLIGHT"]}
          </div>
          <Button
            text={"New Payment"}
            onClick={() => {
              setDestinationType(0);
              setDestination("");
              setDestState(ProgressStepState.active);
              setConfirmState(ProgressStepState.disabled);
              setProcessState(ProgressStepState.disabled);
              setStepIndex(0);
              setResponses([]);
            }}
            buttonColor={buttonColor.secondary}
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewPaymentModal;
