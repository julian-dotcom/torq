import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./newInvoice.module.scss";
import classNames from "classnames";
import { useEffect } from "react";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { format } from "d3";
import { NewInvoiceResponse } from "./newInvoiceTypes";
import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";
import Note, { NoteType } from "features/note/Note";
import {
  DetailsContainer,
  DetailsRow,
  DetailsRowLinkAndCopy,
} from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";

const f = format(",.0f");

type NewInvoiceResponseProps = {
  response?: {
    data: NewInvoiceResponse | undefined;
    isUninitialized: boolean;
    isSuccess?: boolean;
    isLoading?: boolean;
    isError?: boolean;
    error?: FetchBaseQueryError | SerializedError;
  };
  amount: number;
  selectedNodeId: number;
  setDoneState: (state: ProgressStepState) => void;
  clearFlow: () => void;
};

export function NewInvoiceResponseStep(props: NewInvoiceResponseProps) {
  useEffect(() => {
    if (props.response?.isSuccess) {
      props.setDoneState(ProgressStepState.completed);
    } else if (props.response?.isError) {
      props.setDoneState(ProgressStepState.error);
    }
  }, [props.response]);

  return (
    <ProgressTabContainer>
      {props.response?.isSuccess && (
        <div className={classNames(styles.amountWrapper)}>
          <div className={styles.invoiceStatusMessage}>{`Invoice created`}</div>
          <div className={styles.amountMessage}>{`${f(props.amount)} sat`}</div>
        </div>
      )}
      <StatusIcon state={props.response?.isSuccess ? "success" : props.response?.isError ? "error" : "processing"} />
      {props.response?.isError && (
        <Note title={"Error"} noteType={NoteType.error}>
          {"Failed to create invoice"}
        </Note>
      )}
      {props.response?.isSuccess && (
        <div>
          <DetailsContainer>
            <DetailsRow label={"Amount"}>{props.amount}</DetailsRow>
          </DetailsContainer>
          <DetailsContainer>
            <DetailsRowLinkAndCopy label={"Invoice:"} copy={props.response?.data?.paymentRequest}>
              {props.response?.data?.paymentRequest}
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
              props.clearFlow();
            }}
            buttonColor={buttonColor.subtle}
          />
        }
      />
    </ProgressTabContainer>
  );
}
