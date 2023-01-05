import classNames from "classnames";
import { MutableRefObject, useRef } from "react";
import styles from "./socket_dot.module.scss";

type SocketDotProps = {
  x: number;
};

export default function SocketDot(props: SocketDotProps) {
  const connectorRef = useRef() as MutableRefObject<HTMLDivElement>;
  return (
    <div className={classNames(styles.nodeSocket, styles.socket)} ref={connectorRef}>
      <div className={styles.socketDot} />
    </div>
  );
}
