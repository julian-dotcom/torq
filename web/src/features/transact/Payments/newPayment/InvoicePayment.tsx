import { Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import { useState } from "react";
import TextInput from "features/forms/TextInput";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import { SectionContainer } from "features/section/SectionContainer";
import NumberFormat from "react-number-format";
import styles from "./newPayments.module.scss";
import { PaymentType, PaymentTypeLabel } from "./types";
import { DecodedInvoice } from "types/api";
import { SendJsonMessage } from "react-use-websocket/dist/lib/types";

type InvoicePaymentProps = {
  selectedNodeId: number;
  decodedInvoice: DecodedInvoice;
  destinationType: PaymentType;
  destination: string;
  sendJsonMessage: SendJsonMessage;
  setStepIndex: (index: number) => void;
  setDestState: (state: ProgressStepState) => void;
  setConfirmState: (state: ProgressStepState) => void;
  setProcessState: (state: ProgressStepState) => void;
};
const DefualtTimeoutSeconds = 60;

export default function InvoicePayment(props: InvoicePaymentProps) {
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);
  const [amountSat, setAmountSat] = useState<number | undefined>(undefined);
  const [feeLimit, setFeeLimit] = useState<number | undefined>(
    Math.floor(props.decodedInvoice.valueMsat / 1000000) || 100
  );
  const [timeOutSecs, setTimeOutSecs] = useState(DefualtTimeoutSeconds);

  function lnAmountField() {
    if (props.decodedInvoice.valueMsat !== 0) {
      return props.decodedInvoice.valueMsat / 1000;
    }

    return (
      <NumberFormat
        className={styles.amountInput}
        datatype={"number"}
        value={amountSat}
        placeholder={"0 sat"}
        onValueChange={(values) => setAmountSat(parseInt(values.value))}
        thousandSeparator=","
        suffix={" sat"}
      />
    );
  }

  return (
    <ProgressTabContainer>
      <div className={styles.amountWrapper}>
        {props.destinationType && (
          <span className={styles.destinationType}>{PaymentTypeLabel[props.destinationType] + " Detected"}</span>
        )}
        <div className={styles.amount}>{lnAmountField()}</div>
        <div className={styles.label}>To</div>
        <div className={styles.destinationPreview}>{props.decodedInvoice.nodeAlias}</div>
      </div>
      <SectionContainer
        title={"Advanced Options"}
        icon={OptionsIcon}
        expanded={expandAdvancedOptions}
        handleToggle={() => setExpandAdvancedOptions(!expandAdvancedOptions)}
      >
        <TextInput
          label={"Fee limit"}
          inputType={"number"}
          value={feeLimit}
          onChange={(e) => {
            setFeeLimit(e as number);
          }}
        />
        <TextInput
          label={"Timeout (Seconds)"}
          value={timeOutSecs}
          onChange={(e) => setTimeOutSecs(parseInt(e as string))}
        />
      </SectionContainer>

      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        leftChildren={
          <Button
            text={"Back"}
            onClick={() => {
              props.setStepIndex(0);
              props.setDestState(ProgressStepState.completed);
              props.setConfirmState(ProgressStepState.active);
            }}
            buttonColor={buttonColor.ghost}
          />
        }
        rightChildren={
          <Button
            text={"Confirm"}
            onClick={() => {
              props.sendJsonMessage({
                reqId: "randId",
                type: "newPayment",
                NewPaymentRequest: {
                  nodeId: props.selectedNodeId,
                  // If the destination is not a pubkey, use it as an invoice
                  invoice: props.destination,
                  // If the destination is a pubkey send it as a dest input
                  // dest: destination.match(LightningNodePubkeyRegEx) ? destination : undefined,
                  amtMSat: amountSat ? amountSat * 1000 : undefined, // 1 sat = 1000 msat
                  timeOutSecs: timeOutSecs,
                  feeLimitMsat: feeLimit ? feeLimit * 1000 : 1000 * 1000, // 1 sat = 1000 msat
                  allowSelfPayment: true, //allowSelfPayment
                },
              });
              props.setStepIndex(2);
              props.setConfirmState(ProgressStepState.completed);
              props.setProcessState(ProgressStepState.processing);
            }}
            buttonColor={buttonColor.green}
          />
        }
      />
    </ProgressTabContainer>
  );
}
