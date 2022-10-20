import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as ErrorIcon,
} from "@fluentui/react-icons";
import styles from "./statusIcon.module.scss";
import classNames from "classnames";
import { ReactNode } from "react";

type NewInvoiceResponseProps = {
  errorMessage?: string;
  state: "processing" | "success" | "error";
};
const stateClasses = new Map<string, string>([
  ["processing", styles.processing],
  ["error", styles.error],
  ["success", styles.success],
]);

const stateIcons = new Map<string, ReactNode>([
  ["processing", <ProcessingIcon key={"processingIcon"} />],
  ["error", <ErrorIcon key={"errorIcon"} />],
  ["success", <SuccessIcon key={"successIcon"} />],
]);

export function StatusIcon(props: NewInvoiceResponseProps) {
  return (
    <div className={classNames(styles.statusWrapper, stateClasses.get(props.state) || styles.processing)}>
      {stateIcons.get(props.state) || <ProcessingIcon />}
    </div>
  );
}
