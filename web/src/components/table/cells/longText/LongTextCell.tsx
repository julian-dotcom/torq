import React from "react";
import classNames from "classnames";
import { Copy16Regular, Eye16Regular } from "@fluentui/react-icons";
import ToastContext from "features/toast/context";
import { copyToClipboard } from "utils/copyToClipboard";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./long_text_cell.module.scss";
import { toastCategory } from "features/toast/Toasts";

export type TextCellProps = {
  current: string;
  link?: string;
  copyText?: string;
  className?: string;
  totalCell?: boolean;
};

const LongTextCell = (props: TextCellProps) => {
  const toastRef = React.useContext(ToastContext);

  const copyText = () => {
    copyToClipboard(props.copyText || "");
    toastRef?.current?.addToast("Copied to clipboard", toastCategory.success);
  };

  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.textCell,
        { [cellStyles.totalCell]: props.totalCell },
        props.className
      )}
    >
      {props.current && (
        <div className={classNames(styles.action, styles.view)}>
          <Eye16Regular />
          <span className={classNames(styles.content)}>{props.current}</span>
        </div>
      )}
      {props.copyText && (
        <button className={classNames(styles.action, styles.copy)} onClick={copyText}>
          <Copy16Regular />
          Copy
        </button>
      )}
      {props.link && (
        <a href={props.link} className={classNames(styles.action, styles.link)} target="_blank" rel="noreferrer">
          <Eye16Regular />
          Link
        </a>
      )}
    </div>
  );
};

export default LongTextCell;
