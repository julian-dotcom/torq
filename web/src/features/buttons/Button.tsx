import { MouseEventHandler } from "react";
import classNames from "classnames";
import styles from "./button.module.scss";

function DefaultButton(props: {
  text: string;
  icon?: any;
  onClick?: MouseEventHandler<HTMLButtonElement> | undefined;
  className?: string;
  isOpen?: boolean;
}) {
  return (
    <div
      className={classNames(styles.button, props.className, {
        [styles.open]: props.isOpen,
      })}
      //@ts-expect-error
      onClick={props.onClick}
    >
      {props.icon && <div className="icon">{props.icon}</div>}
      <div className="text">{props.text}</div>
    </div>
  );
}

export default DefaultButton;
