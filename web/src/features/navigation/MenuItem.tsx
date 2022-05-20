import styles from "./menu_item.module.scss";
import classNames from "classnames";
import { Link } from "react-router-dom";

function MenuItem(props: {
  text: string;
  selected?: boolean;
  icon?: any;
  children?: any;
  actions?: any;
  routeTo?: string;
}) {
  let TitleComponent = function (routeTo?: string) {
    if (routeTo) {
      return (
        <Link
          to={routeTo}
          className={classNames(styles.title, {
            [styles.noIcon]: !!props.icon,
          })}
        >
          <div className={styles.icon}>{props.icon}</div>
          <div className={styles.text}>{props.text}</div>
        </Link>
      );
    } else {
      return (
        <div
          className={classNames(styles.title, {
            [styles.noIcon]: !!props.icon,
          })}
        >
          <div className={styles.icon}>{props.icon}</div>
          <div className={styles.text}>{props.text}</div>
        </div>
      );
    }
  };

  return (
    <div
      className={classNames(styles.item, { [styles.selected]: props.selected })}
    >
      <div className={styles.contentWrapper}>
        {TitleComponent(props.routeTo)}

        {props.actions && (
          <div className={styles.actions}>
            <div className={classNames(styles.icon, styles.action)}>
              {props.actions}
            </div>
          </div>
        )}
      </div>

      {props.children && (
        <div className={classNames(styles.item, styles.sub)}>
          {props.children}
        </div>
      )}
    </div>
  );
}

export default MenuItem;
