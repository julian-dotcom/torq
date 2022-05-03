import styles from "./switch.module.scss";
import classNames from "classnames";

export default function Switch({ label }: { label: string }) {
  return (
    <label className={styles.switch}>
      <span className={styles.innerSwitch}>
        <input type="checkbox" />
        <span className={classNames(styles.slider, styles.round)}></span>
      </span>
      <div>{label}</div>
    </label>
  );
}
