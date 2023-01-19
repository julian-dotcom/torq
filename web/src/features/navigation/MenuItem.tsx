import React from "react";
import classNames from "classnames";
import { NavLink, useMatch, useResolvedPath } from "react-router-dom";
import { useLocation } from "react-router-dom";
import styles from "./nav.module.scss";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function MenuItem(props: { text: string; icon?: any; routeTo: string; withBackground?: boolean }) {
  const resolvedPath = useResolvedPath(props.routeTo);
  const hasMatch = useMatch({ path: resolvedPath.pathname, end: true });

  const location = useLocation();

  const resolvedClassNames = classNames(styles.title, { [styles.selected]: hasMatch });

  return (
    <div className={classNames(styles.item)}>
      <div className={classNames(styles.contentWrapper)}>
        <NavLink
          to={props.routeTo}
          className={resolvedClassNames}
          state={props.withBackground ? { background: { pathname: location.pathname } } : undefined}
        >
          <div className={styles.icon}>{props.icon}</div>
          <div className={classNames(styles.text)}>{props.text}</div>
        </NavLink>
      </div>
    </div>
  );
}

export default React.memo(MenuItem);
