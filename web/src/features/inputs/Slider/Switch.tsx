import styles from "./switch.module.scss";
import classNames from "classnames";
import React, { HTMLInputTypeAttribute, InputHTMLAttributes } from "react";

export default function Switch({
  label,
  checkboxProps,
  labelPosition,
}: {
  label: string;
  checkboxProps?: InputHTMLAttributes<HTMLInputElement>;
  labelPosition?: "left" | "right";
}) {
  return (
    <label className={styles.switch}>
      <span className={styles.innerSwitch}>
        <input {...checkboxProps} type="checkbox" />
        <span className={classNames(styles.slider, styles.round)}></span>
      </span>
      <div
        className={classNames({
          [styles.left]: labelPosition && labelPosition === "left",
        })}
      >
        {label}
      </div>
    </label>
  );
}
