import classNames from "classnames";
import React from "react";
import styles from "./switch.module.scss";

export enum SwitchSize {
  normal = "normal",
  small = "small",
  tiny = "tiny",
}

const SwitchSizeClasses = new Map<SwitchSize, string>([
  [SwitchSize.normal, styles.normal],
  [SwitchSize.small, styles.small],
  [SwitchSize.tiny, styles.tiny],
]);

export interface SwitchProps
  extends React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> {
  sizeVariant?: SwitchSize;
  label: string;
}

export default function Switch({ label, sizeVariant, ...rest }: SwitchProps) {
  const sizeClass = SwitchSizeClasses.get(sizeVariant || SwitchSize.normal);
  return (
    <label className={classNames(styles.switchWrapper, sizeClass)}>
      <span className={styles.innerSwitch}>
        <input {...rest} type="checkbox" />
        <span className={classNames(styles.slider, styles.round)}></span>
      </span>
      <div>{label}</div>
    </label>
  );
}
