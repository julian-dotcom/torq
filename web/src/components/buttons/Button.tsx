import classNames from "classnames";
import styles from "./button.module.scss";
import { SizeVariant, ColorVariant, GetSizeClass, GetColorClass } from "./buttonVariants";
import { ReactNode } from "react";
import { Link, LinkProps } from "react-router-dom";
// Exporting them here again so that we don't have to import from two different places
export { SizeVariant, ColorVariant } from "./buttonVariants";

export enum ButtonPosition {
  left,
  right,
  center,
  fullWidth,
}

const ButtonPositionClass = new Map([
  [ButtonPosition.left, styles.positionLeft],
  [ButtonPosition.right, styles.positionRight],
  [ButtonPosition.center, styles.positionCenter],
  [ButtonPosition.fullWidth, styles.positionFullWidth],
]);

export type ButtonProps = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon?: any;
  buttonColor?: ColorVariant;
  buttonPosition?: ButtonPosition;
  buttonSize?: SizeVariant;
  children?: ReactNode;
  hideMobileText?: boolean;
};

export default function Button({
  icon,
  buttonColor,
  buttonPosition,
  buttonSize,
  children,
  ...buttonProps
}: React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> & ButtonProps) {
  const color = buttonProps.disabled ? ColorVariant.disabled : buttonColor;

  return (
    <button
      {...buttonProps}
      className={classNames(
        styles.button,
        GetColorClass(color),
        ButtonPositionClass.get(buttonPosition || ButtonPosition.left),
        GetSizeClass(buttonSize),
        { [styles.collapseTablet]: buttonProps.hideMobileText },
        buttonProps.className
      )}
    >
      {icon && <span>{icon}</span>}
      {children && <span className={styles.text}>{children}</span>}
    </button>
  );
}

export function LinkButton({
  icon,
  buttonColor,
  buttonPosition,
  buttonSize,
  children,
  ...buttonProps
}: LinkProps & ButtonProps) {
  return (
    <Link
      {...buttonProps}
      className={classNames(
        styles.button,
        GetColorClass(buttonColor),
        ButtonPositionClass.get(buttonPosition || ButtonPosition.left),
        GetSizeClass(buttonSize),
        buttonProps.className
      )}
    >
      {icon && <span>{icon}</span>}
      {children && <span className={styles.text}>{children}</span>}
    </Link>
  );
}

export function ButtonWrapper(props: {
  leftChildren?: Array<React.ReactNode> | React.ReactNode;
  rightChildren?: Array<React.ReactNode> | React.ReactNode;
  className?: string;
}) {
  return (
    <div className={classNames(styles.buttonWrapper, props.className)}>
      <div className={styles.leftButtonContainer}>{props.leftChildren}</div>
      <div className={styles.rightButtonContainer}>{props.rightChildren}</div>
    </div>
  );
}
