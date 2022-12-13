import styles from "components/table/cells/cell.module.scss";
import React from "react";
import { format } from "d3";
import classNames from "classnames";
import { ArrowStepIn16Regular as ArrowIn, ArrowStepOut16Regular as ArrowOut } from "@fluentui/react-icons";

interface numericDoubleCell {
  local: number;
  remote: number;
  suffix?: string;
  className?: string;
  totalCell?: boolean;
}
const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function NumericDoubleCell({ local, remote, suffix, className, totalCell }: numericDoubleCell) {
  const localValue = local % 1 != 0 ? formatterDetailed(local) : formatter(local);
  const remoteValue = remote % 1 != 0 ? formatterDetailed(remote) : formatter(remote);
  return (
    <div className={classNames(styles.cell, styles.numericCell, className)}>
      {!totalCell && (
        <>
          <div className={styles.local}>
            {localValue} {suffix} <ArrowOut className={styles.inboundIcon} />
          </div>
          <div className={classNames(styles.remote, styles.outbound)}>
            {remoteValue} {suffix} <ArrowIn className={styles.outboundIcon} />
          </div>
        </>
      )}
    </div>
  );
}

const NumericDoubleCellMemo = React.memo(NumericDoubleCell);
export default NumericDoubleCellMemo;
