import React from "react";
import classNames from "classnames";
import { Link, useMatch, useResolvedPath } from "react-router-dom";
import { useLocation } from "react-router-dom";
import styles from "./nav.module.scss";

function MenuItem(props: {
  text: string;
  icon?: JSX.Element;
  routeTo: string;
  withBackground?: boolean;
  onClick?: (e: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => void;
  intercomTarget?: string;
}) {
  const resolvedPath = useResolvedPath(props.routeTo);
  const hasMatch = useMatch({ path: resolvedPath.pathname, end: true });

  const location = useLocation();

  const resolvedClassNames = classNames(styles.title, { [styles.selected]: hasMatch });

  return (
    <div className={classNames(styles.item)} data-intercom-target={props.intercomTarget}>
      <div className={classNames(styles.contentWrapper)}>
        <Link
          onClick={props.onClick}
          to={props.routeTo}
          className={resolvedClassNames}
          state={props.withBackground ? { background: { pathname: location.pathname } } : undefined}
        >
          <div className={styles.icon}>{props.icon}</div>
          <div className={classNames(styles.text)}>{props.text}</div>
        </Link>
      </div>
    </div>
  );
}

export default React.memo(MenuItem);
