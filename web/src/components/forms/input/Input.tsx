import React from "react";
import classNames from "classnames";
import NumberFormat, { NumberFormatProps } from "react-number-format";
import {
  WarningRegular as WarningIcon,
  ErrorCircleRegular as ErrorIcon,
  QuestionCircle16Regular as HelpIcon,
} from "@fluentui/react-icons";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import styles from "./textInput.module.scss";
import labelStyles from "components/forms/label/label.module.scss";
import { BasicInputType } from "components/forms/formTypes";

export type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> &
  BasicInputType;

export type FormattedInputProps = {
  label?: string;
  formatted?: boolean;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
  helpText?: string;
} & NumberFormatProps;

function Input({
  label,
  formatted,
  sizeVariant,
  colorVariant,
  leftIcon,
  errorText,
  warningText,
  helpText,
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
  if (inputProps.disabled === true) {
    inputColorClass = GetColorClass(InputColorVaraint.disabled);
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
      {label && (
        <div className={labelStyles.labelWrapper}>
          <label htmlFor={inputProps.id || inputId} className={styles.label} title={"Something"}>
            {label}
          </label>
          {/* Create a div with a circled question mark icon with a data label named data-title */}
          {helpText && (
            <div className={labelStyles.tooltip}>
              <HelpIcon />
              <div className={labelStyles.tooltipTextWrapper}>
                <div className={labelStyles.tooltipTextContainer}>
                  <div className={labelStyles.tooltipHeader}>{label}</div>
                  <div className={labelStyles.tooltipText}>{helpText}</div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
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
