import styles from "./switch.module.scss";
import classNames from "classnames";
import React, { HTMLInputTypeAttribute, InputHTMLAttributes } from "react";

export default function Switch({
  label,
  checkboxProps,
}: {
  label: string;
  checkboxProps?: InputHTMLAttributes<HTMLInputElement>;
}) {
  return (
    <label className={styles.switch}>
      <span className={styles.innerSwitch}>
        <input {...checkboxProps} type="checkbox" />
        <span className={classNames(styles.slider, styles.round)}></span>
      </span>
      <div>{label}</div>
    </label>
  );
}
