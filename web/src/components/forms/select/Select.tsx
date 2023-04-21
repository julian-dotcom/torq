import ReactSelect, {
  Props,
  components,
  MultiValueRemoveProps,
  GroupBase,
  ClearIndicatorProps,
  DropdownIndicatorProps,
  StylesConfig,
} from "react-select";
import {
  ChevronDown16Regular as ChevronDownIcon,
  Dismiss12Regular as MultiValueRemoveIcon,
  Dismiss16Regular as ClearIcon,
  WarningRegular as WarningIcon,
  ErrorCircleRegular as ErrorIcon,
} from "@fluentui/react-icons";
import styles from "./select.module.scss";
import classNames from "classnames";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "../variants";
import { useId } from "react";

export type SelectOptionType = { value: string; label: string };

const customStyles: StylesConfig<unknown, boolean, GroupBase<unknown>> = {
  indicatorSeparator: () => {
    return {};
  },
  control: (provided, state) => ({
    ...provided,
    borderRadius: 4,
    backgroundColor: state.isFocused ? "var(--input-focus-background)" : "var(--input-default-background)",
    borderColor: state.isFocused ? "var(--input-focus-border-color)" : "transparent",
    boxShadow: "none",
    "&:hover": {
      backgroundColor: state.isFocused ? "var(--input-focus-background)" : "var(--input-hover-background)",
      boxShadow: "none",
    },
    fontSize: "var(--input-font-size)",
    minHeight: "var(--input-height)",
  }),
  placeholder: (provided) => {
    return {
      ...provided,
      color: "var(--input-placeholder-color)",
      fontSize: "var(--input-font-size)",
    };
  },
  dropdownIndicator: (provided) => ({
    ...provided,
    color: "var(--input-color)",
    fontSize: "var(--input-font-size)",
    padding: "var(--indicator-padding)",
    alignItems: "flex-start",
  }),
  singleValue: (provided) => ({
    ...provided,
    color: "var(--input-color)",
    fontSize: "var(--input-font-size)",
  }),
  input: (provided) => ({
    ...provided,
    margin: "0",
    padding: "0",
  }),
  option: (provided, state) => ({
    ...provided,
    color: "var(--input-color)",
    fontSize: "var(--input-font-size)",
    background: state.isFocused ? "var(--input-default-background)" : "#ffffff",
    "&:hover": {
      boxShadow: "none",
      backgroundColor: "var(--input-hover-background)",
    },
    borderRadius: "4px",
  }),
  menuList: (provided) => ({
    ...provided,
    borderColor: "transparent",
    boxShadow: "none",
    padding: "8px",
    display: "flex",
    flexDirection: "column",
    gap: "4px",
  }),
  menu: (provided) => ({
    ...provided,
    margin: "8px 4px",
    clip: "initial",
    width: "100%",
    borderColor: "transparent",
    borderRadius: "4px",
    boxShadow: "var(--hover-box-shadow)",
    zIndex: "10",
  }),
};

export type SelectProps = Props & {
  label?: string;
  colorVariant?: InputColorVaraint;
  sizeVariant?: InputSizeVariant;
  warningText?: string;
  errorText?: string;
  intercomTarget?: string;
};

export default function Select({
  label,
  colorVariant,
  sizeVariant,
  warningText,
  errorText,
  intercomTarget,
  ...selectProps
}: SelectProps) {
  const DropdownIndicator = (props: DropdownIndicatorProps) => {
    return (
      <components.DropdownIndicator {...props}>
        <ChevronDownIcon />
      </components.DropdownIndicator>
    );
  };
  const MultiValueRemove = (props: MultiValueRemoveProps<unknown, boolean, GroupBase<unknown>>) => {
    return (
      <components.MultiValueRemove {...props}>
        <MultiValueRemoveIcon />
      </components.MultiValueRemove>
    );
  };
  const ClearIndicator = (props: ClearIndicatorProps<unknown, boolean, GroupBase<unknown>>) => {
    return (
      <components.ClearIndicator {...props}>
        <ClearIcon />
      </components.ClearIndicator>
    );
  };
  const inputId = useId();
  let inputColorClass = GetColorClass(colorVariant);
  if (warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }
  if (selectProps.isDisabled === true) {
    inputColorClass = GetColorClass(InputColorVaraint.disabled);
  }

  return (
    <div
      className={classNames(styles.inputWrapper, GetSizeClass(sizeVariant), inputColorClass)}
      data-intercom-target={intercomTarget}
    >
      {label && (
        <div className={styles.labelWrapper}>
          <label htmlFor={selectProps.id || inputId} className={styles.label}>
            {label}
          </label>
        </div>
      )}
      <ReactSelect
        id={selectProps.id || inputId}
        components={{ DropdownIndicator, MultiValueRemove, ClearIndicator }}
        className={selectProps.className}
        styles={customStyles}
        {...selectProps}
      />
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
