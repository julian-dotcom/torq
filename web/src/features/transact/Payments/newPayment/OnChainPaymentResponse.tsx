import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import { Link20Regular as LinkIcon, Copy20Regular as CopyIcon } from "@fluentui/react-icons";
import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
} from "@fluentui/react-icons";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./newPayments.module.scss";
import { SendOnChainResponse } from "types/api";
import classNames from "classnames";
import { useContext, useEffect } from "react";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { format } from "d3";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";

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
      <div
        className={classNames(styles.paymentResultIconWrapper, {
          [styles.processing]: props.response?.isLoading || props.response?.isUninitialized,
          [styles.failed]: props.response?.isError,
          [styles.success]: props.response?.isSuccess,
        })}
      >
        {props.response?.isUninitialized && <ProcessingIcon />}
        {props.response?.isLoading && <ProcessingIcon />}
        {props.response?.isError && <FailedIcon />}
        {props.response?.isSuccess && <SuccessIcon />}
      </div>
      {props.response?.isSuccess && (
        <div className={styles.txDetailsContainer}>
          <div className={styles.txDetailsRow}>
            <div className={styles.txDetailsLabel}>Destination: </div>
            <div className={styles.txDetailsValue}>{props.destination}</div>
          </div>
          <div className={styles.txDetailsRow}>
            <div className={styles.txDetailsLabel}>Transaction ID: </div>
            <div className={styles.txDetailsButtonsContainer}>
              <div className={styles.txDetailsValue}>{props.response?.data?.txId}</div>
              <div className={styles.txDetailsLink}>
                <a href={mempoolUrl} target="_blank" rel="noreferrer">
                  <LinkIcon />
                </a>
              </div>
              <div className={styles.txDetailsLink}>
                <div
                  onClick={() => {
                    if (props.response?.data?.txId) {
                      navigator.clipboard.writeText(props.response?.data?.txId);
                      toastRef?.current?.addToast("Transaction ID copied to clipboard", toastCategory.success);
                    }
                  }}
                >
                  <CopyIcon />
                </div>
              </div>
            </div>
          </div>
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
