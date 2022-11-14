import styles from "components/table/tableCells/cell.module.scss";
import React from "react";
import classNames from "classnames";

interface textCell {
  current: string | Array<string>;
  className?: string;
}

function TextCell({ current, className }: textCell) {
  let displayCurrent = current;
  if (Array.isArray(current)) {
    displayCurrent = current.join(", ");
  }
  return (
    <div className={classNames(styles.cell, styles.alignLeft, styles.TextCell, className)}>
      <div className={styles.current}>{displayCurrent}</div>
    </div>
  );
}

const TextCellMemo = React.memo(TextCell);
export default TextCellMemo;
