import React from "react";
import styles from "./nav.module.scss";
import classNames from "classnames";
import { ChevronDown20Regular as ExpandIcon, LineHorizontal120Regular as CollapseIcon } from "@fluentui/react-icons";

type NavigationProps = {
  text: string;
  collapsed?: boolean;
  intercomTarget?: string;
  children: React.ReactNode;
};

function NavCategory(props: NavigationProps) {
  const icon = props.collapsed ? <ExpandIcon /> : <CollapseIcon />;
  return (
    <div className={classNames(styles.navCategory)} data-intercom-target={props.intercomTarget}>
      <div
        className={classNames(
          styles.NavCategoryTitleContainer,
          styles.navCollapsedCategoryTitle,
          styles.navCategoryTitle
        )}
      >
        <CollapseIcon />
      </div>
      <div className={classNames(styles.NavCategoryTitleContainer)}>
        <div className={classNames(styles.icon)}>{icon}</div>
        <div className={classNames(styles.navCategoryTitle)}>{props.text}</div>
      </div>
      {!props.collapsed && <div className={styles.menuItemWrapper}>{props.children}</div>}
    </div>
  );
}

export default React.memo(NavCategory);
