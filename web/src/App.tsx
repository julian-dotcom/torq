import React, { useEffect } from "react";

import styles from "./app.module.scss";
import ToastContext from "./features/toast/context";
import Toasts, { addToastHandle } from "./features/toast/Toasts";
import Router from "./Router";
import useTranslations from "./services/i18n/useTranslations";

function App() {
  console.log("App render");
  const { init, status: i18nStatus } = useTranslations();
  const toastRef = React.useRef<addToastHandle>();

  useEffect(() => {
    init();
  }, []);

  return i18nStatus === "loading" ? (
    <p>Loading...</p>
  ) : (
    <ToastContext.Provider value={toastRef}>
      <div className={styles.app}>
        <Toasts ref={toastRef} />
        <Router />
      </div>
    </ToastContext.Provider>
  );
}

export default App;
