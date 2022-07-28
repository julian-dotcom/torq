import styles from "./nav.module.scss";
import classNames from "classnames";
import {
  ChevronDown20Regular as ExpandIcon,
  ChevronLeft20Regular as ExpandedIcon,
  LineHorizontal120Regular as CollapseIcon,
} from "@fluentui/react-icons";

function NavCategory(props: { text: string; collapsed?: boolean; children: React.ReactNode }) {
  const icon = props.collapsed ? <ExpandIcon /> : <CollapseIcon />;
  return (
    <div className={classNames(styles.navCategory)}>
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
        <div className={classNames(styles.navCategoryTitle)}>{props.text}</div>
        <div className={classNames(styles.icon)}>{icon}</div>
      </div>
      {!props.collapsed && <div>{props.children}</div>}
    </div>
  );
}

export default NavCategory;
