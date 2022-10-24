import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";

import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./newPayments.module.scss";
import { SendOnChainResponse } from "types/api";
import classNames from "classnames";
import { useContext, useEffect } from "react";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { format } from "d3";
import ToastContext from "features/toast/context";
import {
  DetailsContainer,
  DetailsRow,
  DetailsRowLinkAndCopy,
} from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";
import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";

const f = format(",.0f");

type OnChainPaymentResponseProps = {
  response?: {
    data: SendOnChainResponse;
    isUninitialized: boolean;
    isSuccess?: boolean;
    isLoading?: boolean;
    isError?: boolean;
    error?: FetchBaseQueryError | SerializedError;
  };
  amount: number;
  selectedLocalNode: number;
  setProcessState: (state: ProgressStepState) => void;
  destination: string;
  clearPaymentFlow: () => void;
};

export function OnChainPaymentResponse(props: OnChainPaymentResponseProps) {
  const toastRef = useContext(ToastContext);
  useEffect(() => {
    if (props.response?.isSuccess) {
      props.setProcessState(ProgressStepState.completed);
    } else if (props.response?.isError) {
      props.setProcessState(ProgressStepState.error);
    }
  }, [props.response]);

  const mempoolUrl = `https://mempool.space/tx/${props.response?.data?.txId}`;

  return (
    <ProgressTabContainer>
      {props.response?.isSuccess && (
        <div className={classNames(styles.paymentStatusMessage)}>
          <div className={styles.amountPaid}>{`${f(props.amount)} sat`}</div>
          <div className={styles.amountPaidText}>{`On-chain payment broadcasted`}</div>
        </div>
      )}
      {props.response?.isError && (
        <div className={classNames(styles.paymentStatusMessage)}>
          <div className={styles.amountPaidText}>{`Failed on-chain payment`}</div>
        </div>
      )}
      <StatusIcon state={props.response?.isSuccess ? "success" : props.response?.isError ? "error" : "processing"} />
      {props.response?.isSuccess && (
        <div>
          <DetailsContainer>
            <DetailsRow label={"Destination:"}>{props.destination}</DetailsRow>
          </DetailsContainer>
          <DetailsContainer>
            <DetailsRowLinkAndCopy label={"Transaction ID:"} copy={props.response?.data?.txId} link={mempoolUrl}>
              {props.response?.data?.txId}
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
