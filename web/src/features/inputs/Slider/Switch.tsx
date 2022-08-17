import styles from "./switch.module.scss";
import classNames from "classnames";
import React, { HTMLInputTypeAttribute, InputHTMLAttributes } from "react";

export default function Switch({
  label,
  checked,
  onChange,
  labelPosition,
}: {
  label: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  labelPosition?: "left" | "right";
}) {
  const handleChange = (value: React.ChangeEvent<HTMLInputElement>) => {
    if (onChange) {
      onChange(value.target.checked);
    }
  };
  return (
    <label className={styles.switch}>
      <span className={styles.innerSwitch}>
        <input checked={checked} onChange={handleChange} type="checkbox" />
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
