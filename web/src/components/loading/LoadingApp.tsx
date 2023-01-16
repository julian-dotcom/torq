import styles from "./loading.module.scss";
import { ArrowSyncFilled as ProcessingIcon } from "@fluentui/react-icons";

export default function LoadingApp() {
  return (
    <div className={styles.loadingApp}>
      <div className={styles.loadingIconWrapper}>
        <ProcessingIcon className={styles.loadingIcon} />
      </div>
      <div className={styles.loadingText}>{"Loading..."}</div>
    </div>
  );
}
