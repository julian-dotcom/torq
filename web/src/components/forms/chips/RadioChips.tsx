import classNames from "classnames";
import styles from "./radio_chips.module.scss";
import labelStyles from "components/forms/label/label.module.scss";
import {
  // WarningRegular as WarningIcon,
  // ErrorCircleRegular as ErrorIcon,
  QuestionCircle16Regular as HelpIcon,
} from "@fluentui/react-icons";
import { GetColorClass, InputColorVaraint, InputSizeVariant } from "components/forms/variants";
import { GetSizeClass } from "../input/variants";
import { DetailedHTMLProps, InputHTMLAttributes } from "react";

export type RadioChipsProps = {
  label?: string;
  groupName: string;
  options: Array<DetailedHTMLProps<InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & { label: string }>;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  editingDisabled?: boolean | false;
  helpText?: string;
  vertical?: boolean;
  intercomTarget?: string;
};

export default function RadioChips(props: RadioChipsProps) {
  const sizeClass = GetSizeClass(props.sizeVariant || InputSizeVariant.normal);
  const colorClass = GetColorClass(props.colorVariant || InputColorVaraint.primary);
  return (
    <div
      className={classNames(styles.radioChipsWrapper, sizeClass, colorClass, { [styles.vertical]: props.vertical })}
      data-intercom-target={props.intercomTarget}
    >
      {props.label && (
        <div className={labelStyles.labelWrapper}>
          <label htmlFor="groupName" className={styles.radioChipsLabel}>
            {props.label}
          </label>
          {props.helpText && (
            <div className={labelStyles.tooltip}>
              <HelpIcon />
              <div className={labelStyles.tooltipTextWrapper}>
                <div className={labelStyles.tooltipTextContainer}>
                  <div className={labelStyles.tooltipHeader}>{props.label}</div>
                  <div className={labelStyles.tooltipText}>{props.helpText}</div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
      <div className={styles.radioChipsContainer}>
        {props.options.map((option, index) => {
          const { label, ...inputOptions } = option;
          return (
            <div className={styles.optionWrapper} key={"radio-chip-" + index}>
              <input
                {...inputOptions}
                id={inputOptions.id}
                type="radio"
                className={(styles.radioButton, inputOptions.className)}
                name={props.groupName}
                disabled={props.editingDisabled}
              />
              <label htmlFor={option.id} className={styles.optionLabel}>
                {label}
              </label>
            </div>
          );
        })}
      </div>
    </div>
  );
}
