import styles from "./textInput.module.scss";
import React from "react";
import classNames from "classnames";
import NumberFormat, { NumberFormatProps } from "react-number-format";

export enum InputSize {
  normal = "normal",
  small = "small",
  tiny = "tiny",
}

const InputSizeClasses = new Map<InputSize, string>([
  [InputSize.normal, styles.normal],
  [InputSize.small, styles.small],
  [InputSize.tiny, styles.tiny],
]);

export interface InputProps extends NumberFormatProps {
  label?: string;
  sizeVariant?: InputSize;
  leftIcon?: React.ReactNode;
}

function Input({ label, sizeVariant, leftIcon, ...inputProps }: InputProps) {
  const inputId = React.useId();

  return (
    <div className={classNames(styles.inputWrapper, InputSizeClasses.get(sizeVariant || InputSize.normal))}>
      {label && (
        <div className={styles.labelWrapper}>
          <label htmlFor={inputId} className={styles.label}>
            {label}
          </label>
        </div>
      )}
      <div className={classNames(styles.inputFieldContainer, { [styles.hasLeftIcon]: !!leftIcon })}>
        {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
        <NumberFormat {...inputProps} className={classNames(styles.input, inputProps.className)} id={inputId} />
      </div>
    </div>
  );
}
export default Input;
