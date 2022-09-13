import React, { useEffect } from "react";
import styles from "./toast.module.css";
import {
  ThumbLike20Regular as SuccessIcon,
  Dismiss20Regular as CloseIcon,
  ErrorCircle20Regular as ErrorIcon,
  Warning20Regular as WarnIcon,
} from "@fluentui/react-icons";
import { TransitionGroup, CSSTransition } from "react-transition-group";
import { v4 as uuidv4 } from "uuid";
import classNames from "classnames";

export enum toastCategory {
  success,
  warn,
  error,
}

export interface toast {
  message: string;
  category: toastCategory;
  uuid: string;
  timeRemaining: number;
}

export type addToastHandle = {
  addToast: (message: string, category: toastCategory) => void;
};

const Toasts = React.forwardRef(function Toasts(_, ref) {
  const [toasts, setToasts] = React.useState([] as toast[]);

  React.useImperativeHandle(ref, () => ({
    addToast(message: string, category: toastCategory) {
      addToast(message, category);
    },
  }));

  function addToast(message: string, category: toastCategory) {
    setToasts([...toasts, { uuid: uuidv4(), message: message, category: category, timeRemaining: 6 }]);
  }

  useEffect(() => {
    const interval = setInterval(() => {
      setToasts((prev) =>
        prev
          .filter((i) => i.timeRemaining > 0)
          .map((item) => {
            return {
              ...item,
              timeRemaining: item.timeRemaining - 1,
            };
          })
      );
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const removeToast = (uuid: string) => {
    setToasts((prev) => prev.filter((i) => i.uuid !== uuid));
  };

  return (
    <div className={styles.toastContainer}>
      <TransitionGroup className="toast-group">
        {toasts.map((toast) => {
          const ref = React.createRef<HTMLDivElement>();
          return (
            <CSSTransition
              nodeRef={ref}
              key={toast.uuid}
              timeout={1000}
              classNames={{
                enter: styles.toastEnter,
                enterActive: styles.toastEnterActive,
                exit: styles.toastExit,
                exitActive: styles.toastExitActive,
              }}
            >
              <div ref={ref} key={toast.uuid + "toast"} className={styles.toast}>
                <div className={styles.icon}>
                  <div
                    className={classNames(
                      styles.iconBackground,
                      {
                        [styles.success]: toast.category === toastCategory.success,
                      },
                      { [styles.warn]: toast.category === toastCategory.warn },
                      { [styles.error]: toast.category === toastCategory.error }
                    )}
                  >
                    {toast.category === toastCategory.success && <SuccessIcon />}
                    {toast.category === toastCategory.warn && <WarnIcon />}
                    {toast.category === toastCategory.error && <ErrorIcon />}
                  </div>
                </div>
                <div>
                  <p className={styles.message}>{toast.message}</p>
                </div>
                <div>
                  <div className={styles.close} onClick={() => removeToast(toast.uuid)}>
                    <CloseIcon />
                  </div>
                </div>
              </div>
            </CSSTransition>
          );
        })}
      </TransitionGroup>
    </div>
  );
});

export default Toasts;
