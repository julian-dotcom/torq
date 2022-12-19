import React from "react";
import styles from "./Modal.module.scss";
import { DismissCircle24Regular as DismissIcon } from "@fluentui/react-icons";
import classNames from "classnames";

interface ModalProps {
  children: React.ReactNode;
  show: boolean;
  title?: string;
  icon?: React.ReactNode;
  onClose: () => void;
}

const Modal = (props: ModalProps) => {
  const handleClose = () => {
    props.onClose();
  };
  return (
    <div className={classNames(styles.modal, { [styles.show]: props.show })}>
      <div className={styles.modalBackdrop} onClick={handleClose} />
      <div className={classNames(styles.content, { [styles.show]: props.show })}>
        <div className={styles.header}>
          {props.icon && <span className={styles.icon}>{props.icon}</span>}
          <span className={styles.title}>{props.title}</span>
          <span className={styles.close} onClick={handleClose}>
            <DismissIcon />
          </span>
        </div>
        <div className={styles.modalBodyWrapper}>
          <div className={styles.modalBody}>{props.children}</div>
        </div>
      </div>
    </div>
  );
};

export default Modal;
