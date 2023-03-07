import React from "react";
import classNames from "classnames";
import {
  WarningRegular as WarningIcon,
  ErrorCircleRegular as ErrorIcon,
  QuestionCircle16Regular as HelpIcon,
} from "@fluentui/react-icons";
import { GetColorClass, GetSizeClass, InputColorVaraint } from "components/forms/input/variants";
import styles from "components/forms/input/textInput.module.scss";
import labelStyles from "components/forms/label/label.module.scss";
import { BasicInputType } from "components/forms/formTypes";

export type InputProps = React.DetailedHTMLProps<
  React.TextareaHTMLAttributes<HTMLTextAreaElement>,
  HTMLTextAreaElement
> &
  Omit<BasicInputType, "leftIcon">;

function TextArea({ label, sizeVariant, colorVariant, errorText, warningText, helpText, ...inputProps }: InputProps) {
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

  return (
    <div className={classNames(styles.inputWrapper, GetSizeClass(sizeVariant), inputColorClass)}>
      {label && (
        <div className={labelStyles.labelWrapper}>
          <label htmlFor={inputProps.id || inputId} className={styles.label} title={"Something"}>
            {label}
          </label>
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
      <div className={classNames(styles.inputFieldContainer)}>
        <textarea
          {...(inputProps as InputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputProps.id || inputId}
        />
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
export default TextArea;
