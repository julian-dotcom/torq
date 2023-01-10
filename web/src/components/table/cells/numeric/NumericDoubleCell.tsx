import styles from "components/table/cells/cell.module.scss";
import React, { ReactNode } from "react";
import { format } from "d3";
import classNames from "classnames";
import { ArrowStepIn16Regular as ArrowIn, ArrowStepOut16Regular as ArrowOut } from "@fluentui/react-icons";

interface numericDoubleCell {
  topValue: number;
  bottomValue: number;
  topIcon?: ReactNode;
  bottomIcon?: ReactNode;
  suffix?: string;
  className?: string;
  totalCell?: boolean;
}
const formatterDetailed = format(",.2f");
const formatter = format(",.0f");

function NumericDoubleCell({
  topValue,
  bottomValue,
  suffix,
  className,
  totalCell,
  topIcon,
  bottomIcon,
}: numericDoubleCell) {
  const localValue = topValue % 1 != 0 ? formatterDetailed(topValue) : formatter(topValue);
  const remoteValue = bottomValue % 1 != 0 ? formatterDetailed(bottomValue) : formatter(bottomValue);
  topIcon = topIcon ? topIcon : <ArrowOut className={styles.outboundIcon} />;
  bottomIcon = bottomIcon ? bottomIcon : <ArrowIn className={styles.inboundIcon} />;
  return (
    <div className={classNames(styles.cell, styles.numericCell, className)}>
      {!totalCell && (
        <>
          <div className={styles.local}>
            {localValue} {suffix} {topIcon}
          </div>
          <div className={classNames(styles.remote, styles.outbound)}>
            {remoteValue} {suffix} {bottomIcon}
          </div>
        </>
      )}
    </div>
  );
}

const NumericDoubleCellMemo = React.memo(NumericDoubleCell);
export default NumericDoubleCellMemo;
