import React from "react";
import styles from "./cell.module.scss";
import classNames from "classnames";
import { format } from "date-fns";

interface DateCellProps {
  value: string | Date;
  className?: string;
}

function DateCell(props: DateCellProps) {
  let date = props.value;
  if (typeof props.value === "string") {
    date = new Date(props.value);
  }
  return (
    <div className={classNames(styles.cell, styles.alignLeft, styles.booleanCell, styles.DateCell, props.className)}>
      {format(date as Date, "yyyy-MM-dd HH:mm:ss")}
    </div>
  );
}

const DateCellMemo = React.memo(DateCell);
export default DateCellMemo;
