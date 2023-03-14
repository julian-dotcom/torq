import classNames from "classnames";
import styles from "./button.module.scss";
import { SizeVariant, ColorVariant, GetSizeClass, GetColorClass } from "./buttonVariants";
import { AnchorHTMLAttributes, DetailedHTMLProps, forwardRef, LegacyRef, ReactNode } from "react";
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
  icon?: ReactNode;
  buttonColor?: ColorVariant;
  buttonPosition?: ButtonPosition;
  buttonSize?: SizeVariant;
  children?: ReactNode;
  hideMobileText?: boolean;
  hideMobile?: boolean;
};

const Button = forwardRef(
  (
    {
      icon,
      buttonColor,
      buttonPosition,
      buttonSize,
      children,
      hideMobileText,
      hideMobile,
      ...buttonProps
    }: React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> & ButtonProps,
    ref: LegacyRef<HTMLButtonElement> | undefined
  ) => {
    const color = buttonProps.disabled ? ColorVariant.disabled : buttonColor;

    return (
      <button
        ref={ref}
        {...buttonProps}
        className={classNames(
          styles.button,
          GetColorClass(color),
          ButtonPositionClass.get(buttonPosition || ButtonPosition.left),
          GetSizeClass(buttonSize),
          { [styles.collapseTablet]: hideMobileText || false, [styles.hideMobile]: hideMobile || false },
          buttonProps.className
        )}
      >
        {icon && <span>{icon}</span>}
        {children && <span className={styles.text}>{children}</span>}
      </button>
    );
  }
);

Button.displayName = "Button";

export default Button;

export function LinkButton({
  icon,
  buttonColor,
  buttonPosition,
  buttonSize,
  children,
  hideMobileText,
  hideMobile,
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
        { [styles.collapseTablet]: hideMobileText || false, [styles.hideMobile]: hideMobile || false },
        buttonProps.className
      )}
    >
      {icon && <span>{icon}</span>}
      {children && <span className={styles.text}>{children}</span>}
    </Link>
  );
}

export function ExternalLinkButton({
  icon,
  buttonColor,
  buttonPosition,
  buttonSize,
  children,
  hideMobileText,
  hideMobile,
  ...buttonProps
}: ButtonProps & DetailedHTMLProps<AnchorHTMLAttributes<HTMLAnchorElement>, HTMLAnchorElement>) {
  return (
    <a
      {...buttonProps}
      className={classNames(
        styles.button,
        GetColorClass(buttonColor),
        ButtonPositionClass.get(buttonPosition || ButtonPosition.left),
        GetSizeClass(buttonSize),
        { [styles.collapseTablet]: hideMobileText || false, [styles.hideMobile]: hideMobile || false },
        buttonProps.className
      )}
    >
      {icon && <span>{icon}</span>}
      {children && <span className={styles.text}>{children}</span>}
    </a>
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
