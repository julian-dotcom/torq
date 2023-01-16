import React, { useEffect } from "react";
import Router from "./Router";
import mixpanel from "mixpanel-browser";
import useTranslations from "./services/i18n/useTranslations";
import ToastContext from "./features/toast/context";
import Toasts, { addToastHandle } from "./features/toast/Toasts";
import styles from "./app.module.scss";
import { useGetSettingsQuery } from "./apiSlice";
import { Network, selectActiveNetwork } from "features/network/networkSlice";
import { useAppSelector } from "./store/hooks";
import LoadingApp from "./components/loading/LoadingApp";

function App() {
  const { init, status: i18nStatus } = useTranslations();
  const toastRef = React.useRef<addToastHandle>();
  const { data: settingsData } = useGetSettingsQuery();

  const activeNetwork = useAppSelector(selectActiveNetwork);

  useEffect(() => {
    init();
    if (process.env.NODE_ENV === "production") {
      mixpanel.init("f08b3b1c4a2fc9e2c7cc014333cc9233", { ip: false });
    } else {
      mixpanel.init("729ace78d0aeb71ba633741d8c92a9ca", { ip: false });
    }
  }, []);

  useEffect(() => {
    if (settingsData) {
      mixpanel.identify(settingsData.torqUuid);
      mixpanel.people.set({
        $opt_out: settingsData.mixpanelOptOut,
      });
      mixpanel.people.set_once({
        $created: new Date().toISOString(),
      });
      mixpanel.register({
        default_date_range: settingsData.defaultDateRange,
        defaultLanguage: settingsData.defaultLanguage,
        weekStartsOn: settingsData.weekStartsOn,
      });
    }
  }, [settingsData]);

  useEffect(() => {
    if (settingsData) {
      mixpanel.register({
        network: Network[activeNetwork],
      });
    }
  }, [activeNetwork]);

  return i18nStatus === "loading" ? (
    <LoadingApp />
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
