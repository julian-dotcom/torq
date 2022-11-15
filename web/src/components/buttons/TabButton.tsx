import classNames from "classnames";
import styles from "./button.module.scss";
import { NavLink } from "react-router-dom";
import { ChevronDown20Regular as ExpandIcon } from "@fluentui/react-icons";

type TabButtonProps = {
  title: string;
  to?: string;
  dropDown?: boolean;
};

function TabButton(props: TabButtonProps) {
  const linkClasses = (p: { isActive: boolean }): string => {
    return classNames(styles.tabButton, { [styles.tabSelected]: p.isActive });
  };
  if (props.to) {
    return (
      <NavLink to={props.to} className={linkClasses}>
        {props.title}
      </NavLink>
    );
  } else {
    return (
      <div className={classNames(styles.tabButton)}>
        {props.title}
        {props.dropDown && (
          <div className={styles.dropDownIcon}>
            <ExpandIcon />
          </div>
        )}
      </div>
    );
  }
}

export default TabButton;
