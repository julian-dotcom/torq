import classNames from "classnames";
import styles from "./button.module.scss";

export enum buttonVariants {
  primary,
  secondary,
  ghost,
  warning,
  green,
}

function Button(props: {
  text: string;
  icon?: any;
  onClick?: Function | undefined;
  className?: string;
  isOpen?: boolean;
  variant: buttonVariants;
  submit?: boolean;
  fullWidth?: boolean;
}) {
  const onClick = () => {
    if (props.onClick) {
      props.onClick();
    }
  };
  return (
    <button
      type={props.submit ? "submit" : "button"}
      className={classNames(styles.button, props.className, {
        [styles.open]: props.isOpen,
        [styles.primary]: props.variant === buttonVariants.primary,
        [styles.secondary]: props.variant === buttonVariants.secondary,
        [styles.ghost]: props.variant === buttonVariants.ghost,
        [styles.warning]: props.variant === buttonVariants.warning,
        [styles.green]: props.variant === buttonVariants.green,
        [styles.wide]: props.fullWidth,
      })}
      onClick={onClick}
    >
      {props.icon && <div className="icon">{props.icon}</div>}
      <div className="text">{props.text}</div>
    </button>
  );
}

export default Button;
