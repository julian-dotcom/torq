import React from "react";
import classNames from "classnames";
import { format } from "date-fns";
import { Clock20Regular as TimeIcon, CalendarLtr20Regular as DateIcon } from "@fluentui/react-icons";
import styles from "./date_cell.module.scss";
import cellStyles from "components/table/tableCells/cell.module.scss";

export interface DateCellProps {
  value?: string | Date;
  className?: string;
}

function DateCell(props: DateCellProps) {
  let date = props.value;
  if (typeof props.value === "string") {
    date = new Date(props.value);
  }
  let displayDate = "";
  let displayTime = "";
  if (date != null && date.valueOf() !== 0) {
    displayDate = format(date as Date, "yyyy-MM-dd");
    displayTime = format(date as Date, "HH:mm:ss");
  } else {
    displayDate = "";
  }

  return (
    <div className={classNames(cellStyles.cell, cellStyles.alignLeft, styles.DateCell, props.className)}>
      <div className={styles.dateRow}>
        {displayDate && <DateIcon />}
        {displayDate}
      </div>
      <div className={styles.timeRow}>
        {displayTime && <TimeIcon />}
        {displayTime}
      </div>
    </div>
  );
}

const DateCellMemo = React.memo(DateCell);
export default DateCellMemo;
