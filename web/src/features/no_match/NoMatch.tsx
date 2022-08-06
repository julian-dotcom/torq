import styles from "./no-match.module.scss";

function NoMatch() {
  return (
    <div className={styles.wrapper}>
      <h1 className={styles.heading}>404 - Page not found</h1>
    </div>
  );
}

export default NoMatch;
