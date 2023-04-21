import React from "react";
import classNames from "classnames";
import styles from "./nav.module.scss";

function MenuItem(props: {
  text: string;
  id?: string;
  icon?: JSX.Element;
  withBackground?: boolean;
  selected?: boolean;
  onClick?: (e: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
  intercomTarget?: string;
}) {
  const resolvedClassNames = classNames(styles.title, styles.menuItemButton, { [styles.selected]: props.selected });
  return (
    <div className={classNames(styles.item)} data-intercom-target={props.intercomTarget}>
      <div className={classNames(styles.contentWrapper)}>
        <div onClick={props.onClick} className={resolvedClassNames} id={props.id}>
          <div className={styles.icon}>{props.icon}</div>
          <div className={classNames(styles.text)}>{props.text}</div>
        </div>
      </div>
    </div>
  );
}

export default React.memo(MenuItem);
