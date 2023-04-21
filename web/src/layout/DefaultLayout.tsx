import { useEffect } from "react";
import { Outlet, useMatch } from "react-router-dom";
import { useAppSelector } from "store/hooks";
import { selectHidden } from "features/navigation/navSlice";
import styles from "./default-layout.module.scss";
import navStyles from "features/navigation/nav.module.scss";
import Navigation from "features/navigation/Navigation";
import TopNavigation from "features/navigation/TopNavigation";
import classNames from "classnames";
import mixpanel from "mixpanel-browser";
import { useGetNodeConfigurationsQuery, useGetServicesQuery, useGetSettingsQuery } from "apiSlice";
import { Network, selectActiveNetwork } from "features/network/networkSlice";
import { useIntercom } from "react-use-intercom";
import { userEvents } from "utils/userEvents";

function DefaultLayout() {
  const { trackEvent } = useIntercom();
  const { track } = userEvents();
  const hidden = useAppSelector(selectHidden);
  const isDashboardPage = useMatch("/");
  const { data: settingsData } = useGetSettingsQuery();
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const { data: servicesData } = useGetServicesQuery();

  useEffect(() => {
    if (process.env.NODE_ENV === "production" && process.env.REACT_APP_E2E_TEST !== "true") {
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

  useEffect(() => {
    if (nodeConfigurations?.length) {
      mixpanel.register({
        nodeCount: nodeConfigurations?.length || 0,
      });
      if (nodeConfigurations?.length === 0) {
        trackEvent("No Node Configured");
      } else {
        console.log("Node Configured");
        trackEvent("Node Configured");
        // check if all data is synced
        const allDataSynced = (nodeConfigurations || []).every((node) => node.status === 1);
        if (allDataSynced) {
          console.log("Node synced");
          track("All Data Synced", { nodeCount: nodeConfigurations?.length || 0 });
        }
      }
    }
  }, [nodeConfigurations?.length, servicesData?.lndServices?.length]);

  return (
    <div
      className={classNames(styles.mainContentWrapper, isDashboardPage ? styles.background : "", {
        [navStyles.navCollapsed]: hidden,
      })}
    >
      <TopNavigation />
      <div className={navStyles.navigationWrapper}>
        <Navigation />
      </div>
      <div className={styles.pageWrapper}>
        {/*<div className="dismiss-navigation-background" onClick={() => dispatch(toggleNav())} />*/}
        <Outlet />
      </div>
    </div>
  );
}

export default DefaultLayout;
