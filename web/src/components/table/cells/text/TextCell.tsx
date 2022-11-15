import styles from "components/table/cells/cell.module.scss";
import React from "react";
import classNames from "classnames";

export type TextCellProps = {
  current: string | Array<string>;
  className?: string;
  total?: boolean;
};

function TextCell(props: TextCellProps) {
  let displayCurrent = props.current;
  if (Array.isArray(props.current)) {
    displayCurrent = props.current.join(", ");
  }
  return (
    <div
      className={classNames(
        styles.cell,
        styles.alignLeft,
        styles.TextCell,
        { [styles.totalCell]: props.total },
        props.className
      )}
    >
      <div className={styles.current}>{displayCurrent}</div>
    </div>
  );
}

const TextCellMemo = React.memo(TextCell);
export default TextCellMemo;
