import styles from "./cell.module.scss";

import React from "react";
import { format } from "d3";
import { toNumber } from "lodash";
const classNames = require("classnames");

interface numericCell {
  current: number;
  previous?: number;
  className?: string;
}
const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function NumericCell({ current, previous, className }: numericCell) {
  return (
    <div className={classNames(styles.cell, styles.numericCell, className)}>
      <div className={styles.current}>{current % 1 != 0 ? formatterDetailed(current) : formatter(current)}</div>
    </div>
  );
}

const NumericCellMemo = React.memo(NumericCell);
export default NumericCellMemo;
