import classNames from "classnames";
import styles from "./radio_chips.module.scss";
import { GetColorClass, InputColorVaraint, InputSizeVariant } from "components/forms/variants";
import { GetSizeClass } from "../input/variants";
import { DetailedHTMLProps, InputHTMLAttributes } from "react";

export type RadioChipsProps = {
  label?: string;
  groupName: string;
  options: Array<DetailedHTMLProps<InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & { label: string }>;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  editingDisabled: boolean;
};

export default function RadioChips(props: RadioChipsProps) {
  const sizeClass = GetSizeClass(props.sizeVariant || InputSizeVariant.normal);
  const colorClass = GetColorClass(props.colorVariant || InputColorVaraint.primary);
  return (
    <div className={classNames(styles.radioChipsWrapper, sizeClass, colorClass)}>
      <label htmlFor="groupName" className={styles.radioChipsLabel}>
        {props.label}
      </label>
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
