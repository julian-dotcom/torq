import React from "react";
import classNames from "classnames";
import styles from "./input_row.module.scss";

type InputRowProps = {
  className?: string;
  children: React.ReactNode;
};

export default function InputRow({ className, children }: InputRowProps) {
  return (
    <div className={classNames(styles.inputRowWrapper, className)}>
      {React.Children.map(children, (child) => {
        return child;
      })}
    </div>
  );
}
