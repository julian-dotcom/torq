import React from "react";
import styles from "components/table/cells/cell.module.scss";
import classNames from "classnames";
import { FluentIconsProps } from "@fluentui/react-icons";

interface EnumCellProps {
  value: string;
  icon?: React.FC<FluentIconsProps>;
  className?: string;
  totalCell?: boolean;
}

function EnumCell(props: EnumCellProps) {
  return (
    <div className={classNames(styles.cell, styles.alignLeft, styles.EnumCell, props.className)}>
      {!props.totalCell && (
        <div className={styles.current}>
          <>
            {props.icon ? props.icon : ""}
            {props.value}
          </>
        </div>
      )}
    </div>
  );
}

const EnumCellMemo = React.memo(EnumCell);
export default EnumCellMemo;
