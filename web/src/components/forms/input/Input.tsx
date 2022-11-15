import styles from "./textInput.module.scss";
import React from "react";
import classNames from "classnames";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import NumberFormat, { NumberFormatProps } from "react-number-format";

export type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & {
  label?: string;
  formatted?: never;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
};
export type FormattedInputProps = NumberFormatProps & {
  label?: string;
  formatted: boolean;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
};

type Only<T, U> = {
  [P in keyof T]: T[P];
} & {
  [P in keyof U]?: never;
};

type Either<T, U> = Only<T, U> | Only<U, T>;

function Input({
  label,
  formatted,
  sizeVariant,
  colorVariant,
  leftIcon,
  errorText,
  warningText,
  ...inputProps
}: InputProps | FormattedInputProps) {
  const inputId = React.useId();
  let inputColorClass = GetColorClass(colorVariant);
  if (warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }

  function renderInput() {
    if (formatted) {
      return (
        <NumberFormat
          {...(inputProps as FormattedInputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputId}
        />
      );
    } else {
      return (
        <input
          {...(inputProps as InputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputId}
        />
      );
    }
  }

  return (
    <div className={classNames(styles.inputWrapper, GetSizeClass(sizeVariant), inputColorClass)}>
      <div className={styles.labelWrapper}>
        <label htmlFor={inputId} className={styles.label}>
          {label}
        </label>
      </div>
      <div className={classNames(styles.inputFieldContainer, { [styles.hasLeftIcon]: !!leftIcon })}>
        {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
        {renderInput()}
      </div>
      <div className={classNames(styles.feedbackWrapper, styles.feedbackError)}>
        <div className={styles.inputErrorIcon}>{errorText}</div>
        <div className={styles.inputErrorText}>{errorText}</div>
      </div>
      <div className={classNames(styles.feedbackWrapper, styles.feedbackWarning)}>
        <div className={styles.inputWarningIcon}>{warningText}</div>
        <div className={styles.inputWarningText}>{warningText}</div>
      </div>
    </div>
  );
}
export default Input;
