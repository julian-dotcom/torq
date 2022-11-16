import React from "react";
import classNames from "classnames";
import NumberFormat, { NumberFormatProps } from "react-number-format";
import { WarningRegular as WarningIcon, ErrorCircleRegular as ErrorIcon } from "@fluentui/react-icons";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import styles from "./textInput.module.scss";

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
          id={inputProps.id || inputId}
        />
      );
    } else {
      return (
        <input
          {...(inputProps as InputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputProps.id || inputId}
        />
      );
    }
  }

  return (
    <div className={classNames(styles.inputWrapper, GetSizeClass(sizeVariant), inputColorClass)}>
      <div className={styles.labelWrapper}>
        <label htmlFor={inputProps.id || inputId} className={styles.label}>
          {label}
        </label>
      </div>
      <div className={classNames(styles.inputFieldContainer, { [styles.hasLeftIcon]: !!leftIcon })}>
        {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
        {renderInput()}
      </div>
      {errorText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackError)}>
          <div className={styles.feedbackIcon}>
            <ErrorIcon />
          </div>
          <div className={styles.feedbackText}>{errorText}</div>
        </div>
      )}
      {warningText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackWarning)}>
          <div className={styles.feedbackIcon}>
            <WarningIcon />
          </div>
          <div className={styles.feedbackText}>{warningText}</div>
        </div>
      )}
    </div>
  );
}
export default Input;
