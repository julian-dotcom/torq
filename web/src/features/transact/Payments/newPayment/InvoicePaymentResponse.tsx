import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
} from "@fluentui/react-icons";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./newPayments.module.scss";
import { DecodedInvoice } from "types/api";
import classNames from "classnames";
import { InvoiceStatusType, NewPaymentResponse } from "../paymentTypes";
import { useEffect, useState } from "react";
import { format } from "d3";
import { StatusIcon } from "../../../templates/popoutPageTemplate/popoutDetails/StatusIcon";
import {
  DetailsContainer,
  DetailsRow,
  DetailsRowLinkAndCopy,
} from "../../../templates/popoutPageTemplate/popoutDetails/PopoutDetails";

const f = format(",.0f");

const paymentStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
};

const paymentStatusClass = {
  IN_FLIGHT: styles.processing,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

type InvoicePaymentResponseProps = {
  responses: Array<NewPaymentResponse>;
  paymentProcessingError: string;
  selectedLocalNode: number;
  decodedInvoice: DecodedInvoice;
  destination: string;
  clearPaymentFlow: () => void;
};

export function InvoicePaymentResponse(props: InvoicePaymentResponseProps) {
  const lastResponse = props.responses[props.responses.length - 1];
  const [status, setStatus] = useState<InvoiceStatusType>("IN_FLIGHT");

  useEffect(() => {
    if (props.paymentProcessingError !== "") {
      setStatus("FAILED");
    } else if (props.responses !== undefined && props.responses.length !== 0) {
      setStatus(lastResponse.status);
    } else {
      setStatus("IN_FLIGHT");
    }
  }, [props.responses]);

  return (
    <ProgressTabContainer>
      {status === "SUCCEEDED" && (
        <div className={classNames(styles.paymentStatusMessage)}>
          <div className={styles.amountPaid}>{`${f(props.decodedInvoice.valueMsat / 1000)} sat`}</div>
          <div className={styles.amountPaidText}>{`Sent to ${props.decodedInvoice.nodeAlias}`}</div>
        </div>
      )}
      <StatusIcon state={status === "SUCCEEDED" ? "success" : status === "FAILED" ? "error" : "processing"} />
      {props.paymentProcessingError && (
        <div className={classNames(styles.paymentStatusMessage)}>{props.paymentProcessingError}</div>
      )}
      {status === "SUCCEEDED" && (
        // <div className={styles.txDetailsContainer}>
        //   <div className={styles.txDetailsRow}>
        //     <div className={styles.txDetailsLabel}>Fee:</div>
        //     <div className={styles.txDetailsValue}>{lastResponse.feePaidMsat} sat</div>
        //   </div>
        // </div>
        <div>
          <DetailsContainer>
            <DetailsRow label={"Fee:"}>{lastResponse.feePaidMsat / 1000}</DetailsRow>
          </DetailsContainer>
          <DetailsContainer>
            <DetailsRowLinkAndCopy label={"Preimage:"} copy={lastResponse.preimage}>
              {lastResponse.preimage}
            </DetailsRowLinkAndCopy>
          </DetailsContainer>
        </div>
      )}
      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        rightChildren={
          <Button
            text={"New payment"}
            onClick={() => {
              props.clearPaymentFlow();
            }}
            buttonColor={buttonColor.subtle}
          />
        }
      />
    </ProgressTabContainer>
  );
}
