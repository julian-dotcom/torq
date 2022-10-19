import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import { Copy20Regular as CopyIcon } from "@fluentui/react-icons";
import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
} from "@fluentui/react-icons";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "features/transact/Payments/newPayment/newPayments.module.scss";
import classNames from "classnames";
import { useContext, useEffect } from "react";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { format } from "d3";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import { NewInvoiceResponse } from "./newInvoiceTypes";

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
  selectedLocalNode: number;
  setDoneState: (state: ProgressStepState) => void;
  clearFlow: () => void;
};

export function NewInvoiceResponseStep(props: NewInvoiceResponseProps) {
  const toastRef = useContext(ToastContext);
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
            <div className={styles.txDetailsLabel}>To node: </div>
            <div className={styles.txDetailsValue}>{props.selectedLocalNode}</div>
          </div>
          <div className={styles.txDetailsRow}>
            <div className={styles.txDetailsLabel}>Invoice: </div>
            <div className={styles.txDetailsButtonsContainer}>
              <div className={styles.txDetailsValue}>{props.response?.data?.paymentRequest}</div>
              <div className={styles.txDetailsLink}>
                <div className={styles.txDetailsLink}>
                  <div
                    onClick={() => {
                      if (props.response?.data?.paymentRequest) {
                        navigator.clipboard.writeText(props.response?.data?.paymentRequest);
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
