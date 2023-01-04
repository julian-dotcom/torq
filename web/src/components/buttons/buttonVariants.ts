import styles from "./button_variations.module.scss";

export enum SizeVariant { // SizeVariant
  large = "large",
  normal = "normal",
  small = "small",
  tiny = "tiny",
}

export enum ColorVariant {
  primary = "primary",
  success = "success",
  error = "error",
  disabled = "disabled",
  warning = "warning",
  accent1 = "accent1",
  accent2 = "accent2",
  accent3 = "accent3",
  ghost = "ghost",
}

export const sizeClasses = new Map<SizeVariant, string>([
  [SizeVariant.large, styles.large],
  [SizeVariant.normal, styles.normal],
  [SizeVariant.small, styles.small],
  [SizeVariant.tiny, styles.tiny],
]);

export const colorVariantClasses = new Map<ColorVariant, string>([
  [ColorVariant.primary, styles.primary],
  [ColorVariant.success, styles.success],
  [ColorVariant.warning, styles.warning],
  [ColorVariant.error, styles.error],
  [ColorVariant.disabled, styles.disabled],
  [ColorVariant.accent1, styles.accent1],
  [ColorVariant.accent2, styles.accent2],
  [ColorVariant.accent3, styles.accent3],
  [ColorVariant.ghost, styles.ghost],
]);

export function GetColorClass(color: ColorVariant | undefined): string {
  return colorVariantClasses.get(color || ColorVariant.primary) || styles.primary;
}

export function GetSizeClass(size: SizeVariant | undefined) {
  return sizeClasses.get(size || SizeVariant.normal) || styles.normal;
}
