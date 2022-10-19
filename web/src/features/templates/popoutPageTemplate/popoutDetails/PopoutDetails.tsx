import { Copy20Regular as CopyIcon } from "@fluentui/react-icons";
import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
} from "@fluentui/react-icons";
import styles from "features/Payments/newPayment/newPayments.module.scss";
import { ReactNode, useContext } from "react";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";

type PopoutDetailsContainer = {
  children?: ReactNode;
};

export function DetailsContainer(props: PopoutDetailsContainer) {
  return <div className={styles.txDetailsContainer}></div>;
}

type DetailsRowProps = {
  label: string;
  children?: ReactNode;
};

export function DetailsRow(props: DetailsRowProps) {
  return (
    <div className={styles.txDetailsRow}>
      <div className={styles.txDetailsLabel}>To node:</div>
      <div className={styles.txDetailsValue}>{props.children}</div>
    </div>
  );
}

type DetailsRowLinkAndCopyProps = {
  label: string;
  children: ReactNode;
  link?: string;
  copy?: string;
};

export function DetailsRowLinkAndCopy(props: DetailsRowLinkAndCopyProps) {
  const toastRef = useContext(ToastContext);

  return (
    <DetailsRow label={props.label}>
      <div className={styles.txDetailsLabel}>To node:</div>
      <div className={styles.txDetailsButtonsContainer}>
        <div className={styles.txDetailsValue}>{props.children}</div>
        {props.copy && (
          <div className={styles.txDetailsLink}>
            <div
              onClick={() => {
                if (props.copy) {
                  navigator.clipboard.writeText(props.copy);
                  toastRef?.current?.addToast("Transaction ID copied to clipboard", toastCategory.success);
                }
              }}
            >
              <CopyIcon />
            </div>
          </div>
        )}
        {props.copy && (
          <div className={styles.txDetailsLink}>
            <div
              onClick={() => {
                if (props.copy) {
                  navigator.clipboard.writeText(props.copy);
                  toastRef?.current?.addToast("Transaction ID copied to clipboard", toastCategory.success);
                }
              }}
            >
              <CopyIcon />
            </div>
          </div>
        )}
      </div>
    </DetailsRow>
  );
}
