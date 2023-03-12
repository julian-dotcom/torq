import { Copy20Regular as CopyIcon, Link20Regular as LinkIcon } from "@fluentui/react-icons";
import styles from "./popoutDetails.module.scss";
import { ReactNode, useContext } from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";

type PopoutDetailsContainer = {
  children?: ReactNode;
};

export function DetailsContainer(props: PopoutDetailsContainer) {
  return <div className={styles.txDetailsContainer}>{props.children}</div>;
}

type DetailsRowProps = {
  label: string;
  children?: ReactNode;
};

export function DetailsRow(props: DetailsRowProps) {
  return (
    <div className={styles.txDetailsRow}>
      <div className={styles.txDetailsLabel}>{props.label}</div>
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
      <div className={styles.txDetailsButtonsContainer}>
        <div className={styles.txDetailsValueButtons}>{props.children}</div>
        {props.copy && (
          <div className={styles.txDetailsLink}>
            <div
              onClick={() => {
                if (props.copy) {
                  navigator.clipboard.writeText(props.copy);
                  toastRef?.current?.addToast("Copied to clipboard", toastCategory.success);
                }
              }}
            >
              <CopyIcon />
            </div>
          </div>
        )}
        {props.link && (
          <div className={styles.txDetailsLink}>
            <a href={props.link} target="_blank" rel="noreferrer">
              <LinkIcon />
            </a>
          </div>
        )}
      </div>
    </DetailsRow>
  );
}
