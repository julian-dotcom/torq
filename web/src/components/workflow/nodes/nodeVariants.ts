import styles from "./node_variants.module.scss";

export enum NodeColorVariant {
  primary = "primary",
  success = "success",
  warning = "warning",
  error = "error",
  accent1 = "accent1",
  accent2 = "accent2",
  accent3 = "accent3",
  disabled = "disabled",
}

export const colorVariantClasses = new Map<NodeColorVariant, string>([
  [NodeColorVariant.primary, styles.primary],
  [NodeColorVariant.success, styles.success],
  [NodeColorVariant.warning, styles.warning],
  [NodeColorVariant.error, styles.error],
  [NodeColorVariant.disabled, styles.disabled],
  [NodeColorVariant.accent1, styles.accent1],
  [NodeColorVariant.accent2, styles.accent2],
  [NodeColorVariant.accent3, styles.accent3],
]);

export function GetColorClass(color: NodeColorVariant | undefined): string {
  return colorVariantClasses.get(color || NodeColorVariant.primary) || styles.primary;
}
