import React from "react";
import classNames from "classnames";
import ToastContext from "features/toast/context";
import { copyToClipboard } from "utils/copyToClipboard";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./text_cell.module.scss";
import { toastCategory } from "features/toast/Toasts";

export type TextCellProps = {
  current: string;
  link?: string;
  copyText?: string;
  className?: string;
  totalCell?: boolean;
};

const TextCell = (props: TextCellProps) => {
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
      {!props.totalCell && (
        <div>
          <span className={classNames(styles.content)}>{props.current}</span>
        </div>
      )}
    </div>
  );
};

const TextCellMemo = React.memo(TextCell);
export default TextCellMemo;
