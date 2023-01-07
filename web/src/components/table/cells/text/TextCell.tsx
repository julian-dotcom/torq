import React from "react";
import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./text_cell.module.scss";

export type TextCellProps = {
  current: string;
  current2?: string;
  link?: string;
  copyText?: string;
  className?: string;
  totalCell?: boolean;
};

const TextCell = (props: TextCellProps) => {
  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.textCell,
        { [cellStyles.totalCell]: props.totalCell },
        props.className
      )}
    >
      {!props.totalCell && (
      <div>
        <div>
          <span className={classNames(styles.content)}>{props.current}</span>
        </div>
        {props.current2 && (
          <div>
            <span className={classNames(styles.content2Row)}>{props.current2}</span>
          </div>
        )}
      </div>
      )}
    </div>
  );
};

const TextCellMemo = React.memo(TextCell);
export default TextCellMemo;
