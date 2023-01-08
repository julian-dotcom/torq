import styles from "./textInput.module.scss";

export enum InputSizeVariant {
  normal = "normal",
  small = "small",
  tiny = "tiny",
  inline = "inline",
}

export enum InputColorVaraint {
  primary = "primary",
  success = "success",
  error = "error",
  warning = "warning",
  accent1 = "accent1",
  accent2 = "accent2",
  accent3 = "accent3",
  disabled = "disabled",
}

export const inputSizeClasses = new Map<InputSizeVariant, string>([
  [InputSizeVariant.normal, styles.normal],
  [InputSizeVariant.small, styles.small],
  [InputSizeVariant.tiny, styles.tiny],
  [InputSizeVariant.inline, styles.inline],
]);

export const colorVaraintClasses = new Map<InputColorVaraint, string>([
  [InputColorVaraint.primary, styles.primary],
  [InputColorVaraint.success, styles.success],
  [InputColorVaraint.warning, styles.warning],
  [InputColorVaraint.error, styles.error],
  [InputColorVaraint.accent1, styles.accent1],
  [InputColorVaraint.accent2, styles.accent2],
  [InputColorVaraint.accent3, styles.accent3],
  [InputColorVaraint.disabled, styles.disabled],
]);

export function GetColorClass(color: InputColorVaraint | undefined): string {
  return colorVaraintClasses.get(color || InputColorVaraint.primary) || styles.primary;
}

export function GetSizeClass(size: InputSizeVariant | undefined) {
  return inputSizeClasses.get(size || InputSizeVariant.normal) || styles.normal;
}
