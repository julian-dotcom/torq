import { ReactElement } from "react";
import styles from "./page.module.scss";

interface pageProps {
  children: ReactElement;
  head?: ReactElement;
}

function Page(props: pageProps) {
  return (
    <div className={styles.page}>
      {props.head && <div className={styles.headContainer}>{props.head}</div>}
      {props.children}
    </div>
  );
}

export default Page;
