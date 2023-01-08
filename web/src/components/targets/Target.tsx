import styles from "./targets.module.scss";
import { Dismiss20Regular as DeleteIcon } from "@fluentui/react-icons";

type TargetProps = {
  title: string;
  details: string;
  icon?: JSX.Element;
  onDeleteTarget?: () => void;
};

export default function Target(props: TargetProps) {
  const { title, details, icon, onDeleteTarget } = props;

  return (
    <div className={styles.targetWrapper}>
      {icon && <div className={styles.icon}>{icon}</div>}
      <div className={styles.targetContent}>
        <div className={styles.title}>{title}</div>
        <div className={styles.details}>{details}</div>
      </div>
      {onDeleteTarget && (
        <div className={styles.delete} onClick={onDeleteTarget}>
          <DeleteIcon />
        </div>
      )}
    </div>
  );
}
