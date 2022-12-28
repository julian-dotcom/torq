import classNames from "classnames";
import styles from "./button.module.scss";
import { SizeVariant, ColorVariant as bc, GetSizeClass, GetColorClass } from "./buttonVariants";
import { ReactNode } from "react";
// Exporting them here again so that we don't have to import from two different places
export { SizeVariant, ColorVariant } from "./buttonVariants";

export enum buttonPosition {
  center,
  fullWidth,
}

const buttonPositionClass = {
  0: styles.positionLeft,
  1: styles.positionRight,
  2: styles.positionCenter,
  3: styles.positionFullWidth,
};

export type ButtonProps = React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> & {
  icon?: any;
  isOpen?: boolean;
  buttonColor: bc;
  buttonPosition?: buttonPosition;
  buttonSize?: SizeVariant;
  children?: ReactNode;
};

export default function Button({
  icon,
  isOpen,
  buttonColor,
  buttonPosition,
  buttonSize,
  children,
  ...buttonProps
}: ButtonProps) {
  const color = buttonProps.disabled ? bc.disabled : buttonColor;

  return (
    <button
      {...buttonProps}
      className={classNames(
        styles.button,
        GetColorClass(color),
        buttonPositionClass[buttonPosition || 0],
        GetSizeClass(buttonSize),
        {
          [styles.open]: isOpen,
        },
        buttonProps.className
      )}
    >
      {icon && <span>{icon}</span>}
      {children && <span className={styles.text}>{children}</span>}
    </button>
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
