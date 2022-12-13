import { format } from "d3";
import React from "react";
import classNames from "classnames";
import styles from "components/table/cells/cell.module.scss";
import barCellStyles from "./barCell.module.scss";

interface barCell {
  current: number;
  max: number;
  className?: string;
  showPercent?: boolean;
}

function formatPercent(num: number) {
  return new Intl.NumberFormat("default", {
    style: "percent",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(num / 100);
}

const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function BarCell({ current, max, className, showPercent }: barCell) {
  const percent = (current || 0) / max;
  let data = current % 1 != 0 ? formatterDetailed(current) : formatter(current);

  if (showPercent) {
    data = formatPercent(current);
  }

  return (
    <div className={classNames(styles.cell, barCellStyles.barCell, className)}>
      <div className={styles.current}>{data}</div>
      <div className={barCellStyles.barWrapper}>
        <div className={barCellStyles.bar} style={{ width: percent * 100 + "%" }} />
        <div className={barCellStyles.totalBar} />
      </div>
    </div>
  );
}

const BarCellMemo = React.memo(BarCell);
export default BarCellMemo;
