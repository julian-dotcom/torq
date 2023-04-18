import React, { useEffect, useState } from "react";
import Router from "./Router";
import useTranslations from "./services/i18n/useTranslations";
import ToastContext from "./features/toast/context";
import Toasts, { addToastHandle } from "./features/toast/Toasts";
import styles from "./app.module.scss";
import navStyles from "features/navigation/nav.module.scss";
import LoadingApp from "./components/loading/LoadingApp";
import { IntercomProvider } from "react-use-intercom";
import classNames from "classnames";
import { polyfillFFDrag } from "utils/drag";

function App() {
  const INTERCOM_APP_ID = "y7n3ouse";
  const { init, status: i18nStatus } = useTranslations();
  const toastRef = React.useRef<addToastHandle>();
  const [helpOpen, setHelpOpen] = useState(false);

  useEffect(() => {
    init();
    polyfillFFDrag();
  }, []);

  function toggleHelp(open: boolean) {
    setHelpOpen(open);
  }

  return i18nStatus === "loading" ? (
    <LoadingApp />
  ) : (
    <IntercomProvider appId={INTERCOM_APP_ID} onHide={() => toggleHelp(false)} onShow={() => toggleHelp(true)}>
      <ToastContext.Provider value={toastRef}>
        <div className={classNames(styles.app, { [navStyles.intercomOpen]: helpOpen })}>
          <Toasts ref={toastRef} />
          <Router />
        </div>
      </ToastContext.Provider>
    </IntercomProvider>
  );
}

export default App;
