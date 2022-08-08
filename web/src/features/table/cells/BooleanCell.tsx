import React from "react";
import styles from "./cell.module.scss";
import classNames from "classnames";
import { CheckmarkCircle24Regular as CheckedIcon, Circle24Regular as UncheckedIcon } from "@fluentui/react-icons";

interface BooleanCellProps {
  value: boolean;
  trueTitle?: string;
  falseTitle?: string;
  className?: string;
}

function BooleanCell(props: BooleanCellProps) {
  const checked = () => {
    return (
      <div>
        <CheckedIcon className={styles.checked} />
      </div>
    );
  };
  const unchecked = () => {
    return (
      <div>
        <UncheckedIcon className={styles.unchecked} />
      </div>
    );
  };
  return (
    <div className={classNames(styles.cell, styles.alignLeft, styles.BooleanCell, props.className)}>
      <div className={styles.current}>{!!props.value ? checked() : unchecked()}</div>
    </div>
  );
}

const BooleanCellMemo = React.memo(BooleanCell);
export default BooleanCellMemo;
