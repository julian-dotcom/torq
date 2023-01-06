import classNames from "classnames";
import { MutableRefObject, useRef } from "react";
import styles from "./socket_dot.module.scss";

export default function SocketDot() {
  const connectorRef = useRef() as MutableRefObject<HTMLDivElement>;
  return (
    <div className={classNames(styles.nodeSocket, styles.socket)} ref={connectorRef}>
      <div className={styles.socketDot} />
    </div>
  );
}
