import styles from "./cell.module.scss";
import React from "react";
import { format } from "d3";
import classNames from "classnames";

interface numericInputCell {
  current: number;
  className?: string;
  onChange?: (value: string | number) => void;
}
const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function NumericInputCell({ current, className, onChange }: numericInputCell) {
  console.log('onChange', onChange)
  const [localValue, setLocalValue] = React.useState<string | number | undefined>("");
  React.useEffect(() => {
    if (current === undefined) {
      return;
    }
    // setLocalValue(current % 1 != 0 ? formatterDetailed(current) : formatter(current));
    setLocalValue(current);
  }, [current]);
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => {
    console.log('e', e.target.value)
    setLocalValue(e.target.value);
    onChange && onChange(e.target.value);
  };
  return (
    <div className={classNames(styles.cell, styles.numericCell, className)}>
      {/* <div className={styles.current}>{current % 1 != 0 ? formatterDetailed(current) : formatter(current)}</div> */}
      <input type={"number"} value={localValue} name={className} onChange={handleChange} className={styles.current} />
    </div>
  );
}

const NumericInputCellMemo = React.memo(NumericInputCell);
export default NumericInputCellMemo;
