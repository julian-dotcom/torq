import styles from "./cell.module.scss";
import { format } from "d3";
import React from "react";
import classNames from "classnames";

interface barCell {
  current: number;
  total: number;
  className?: string;
}
const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

// const percentFormatter = format(".2%");

function BarCell({ current, total, className }: barCell) {
  // const previousPercent = Math.round((current / previous) * Math.random() * 200);
  const percent = (current || 0) / total;
  // const positive: number = Math.round(Math.random());

  return (
    <div className={classNames(styles.cell, styles.barCell, className)}>
      <div className={styles.current}>{current % 1 != 0 ? formatterDetailed(current) : formatter(current)}</div>
      <div className={styles.barWrapper}>
        <div className={styles.bar} style={{ width: percent * 100 + "%" }} />
        <div className={styles.totalBar} />
      </div>
    </div>
  );
}

const BarCellMemo = React.memo(BarCell);
export default BarCellMemo;
