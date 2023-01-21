import classNames from "classnames";
import React from "react";
import styles from "./switch.module.scss";
import { InputColorVaraint, GetColorClass } from "components/forms/variants";

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

export type SwitchProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & {
  sizeVariant?: SwitchSize;
  colorVariant?: InputColorVaraint;
  label: string;
};

export default function Switch({ label, sizeVariant, colorVariant, ...rest }: SwitchProps) {
  const sizeClass = SwitchSizeClasses.get(sizeVariant || SwitchSize.normal);
  const colorClass = GetColorClass(colorVariant || InputColorVaraint.primary);
  return (
    <label className={classNames(styles.switchWrapper, sizeClass, colorClass)}>
      <span className={styles.innerSwitch}>
        <input {...rest} type="checkbox" />
        <span className={classNames(styles.slider, styles.round)}></span>
      </span>
      <div>{label}</div>
    </label>
  );
}
