import React from "react";
import classNames from "classnames";
import styles from "./input_row.module.scss";

type InputRowProps = {
  className?: string;
  children: Array<React.ReactNode>;
};

export default function InputRow({ className, children }: InputRowProps) {
  return (
    <div className={classNames(styles.inputRowWrapper, className)}>
      {(children || []).map((child, index) => {
        return (
          <div className={styles.inputRowItem} key={"input-row-item-" + index}>
            {child}
          </div>
        );
      })}
    </div>
  );
}
