import { ChevronDown20Regular as CollapsedIcon, LineHorizontal120Regular as ExpandedIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import styles from "./targets.module.scss";

type TargetsHeaderProps = {
  title: string;
  icon: React.ReactNode;
  expanded: boolean;
  onCollapse: () => void;
};

export default function TargetsHeader(props: TargetsHeaderProps) {
  return (
    <div className={classNames(styles.header, { [styles.expanded]: props.expanded })} onClick={props.onCollapse}>
      <div className={styles.headerIcon}>{props.icon}</div>
      <div className={styles.headerTitle}>{props.title}</div>
      <div className={styles.collapseIcon}>{props.expanded ? <ExpandedIcon /> : <CollapsedIcon />}</div>
    </div>
  );
}
