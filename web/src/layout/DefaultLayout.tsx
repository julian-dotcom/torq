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
import {
  useGetChannelsQuery,
  useGetNodeConfigurationsQuery,
  useGetNodesInformationByCategoryQuery,
  useGetServicesQuery,
  useGetSettingsQuery,
} from "apiSlice";
import { Network, selectActiveNetwork } from "features/network/networkSlice";
import { userEvents } from "utils/userEvents";
import { channel } from "features/channels/channelsTypes";
import { useIntercom } from "react-use-intercom";

function DefaultLayout() {
  const { boot } = useIntercom();
  const hidden = useAppSelector(selectHidden);
  const isDashboardPage = useMatch("/");
  const { data: settingsData, isSuccess: settingsDataSuccess } = useGetSettingsQuery();
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { data: nodeConfigurations, isSuccess: nodeConfigurationSuccess } = useGetNodeConfigurationsQuery();
  const { data: servicesData, isSuccess: servicesDataSuccess } = useGetServicesQuery();
  const { data: nodes, isSuccess: nodeQueryHasRun } = useGetNodesInformationByCategoryQuery(activeNetwork);

  const { data: channelData } = useGetChannelsQuery<{
    data: Array<channel>;
  }>({ network: activeNetwork });

  useEffect(() => {
    if (settingsData?.torqUuid) {
      boot({
        userId: settingsData?.torqUuid,
        customLauncherSelector: "#intercom-launcher",
        hideDefaultLauncher: true,
      });
    }
  }, [settingsDataSuccess, settingsData?.torqUuid]);

  useEffect(() => {
    if (process.env.NODE_ENV === "production" && process.env.REACT_APP_E2E_TEST !== "true") {
      mixpanel.init("f08b3b1c4a2fc9e2c7cc014333cc9233", { ip: false });
    } else {
      mixpanel.init("729ace78d0aeb71ba633741d8c92a9ca", { ip: false });
    }
  }, []);

  const { track, register } = userEvents();

  useEffect(() => {
    if (settingsData) {
      mixpanel.identify(settingsData.torqUuid);
      mixpanel.people.set({
        $opt_out: settingsData.mixpanelOptOut,
      });
      mixpanel.people.set_once({
        $created: new Date().toISOString(),
      });
      register({
        nodeEnv: process.env.NODE_ENV,
        defaultDateRange: settingsData.defaultDateRange,
        defaultLanguage: settingsData.defaultLanguage,
        weekStartsOn: settingsData.weekStartsOn,
      });
    }
  }, [settingsData]);

  useEffect(() => {
    if (settingsData) {
      register({
        network: Network[activeNetwork],
      });
    }
  }, [activeNetwork]);

  useEffect(() => {
    // Reduce channels data into an object containing channel count and total capacity
    const summary = channelData?.reduce(
      (acc, channel) => {
        acc.channelCount += 1;
        acc.totalCapacity += channel.capacity;
        return acc;
      },
      { channelCount: 0, totalCapacity: 0 }
    );

    register({
      nodeCount: nodeConfigurations?.length || 0,
      channelCount: summary?.channelCount || 0,
      totalCapacity: summary?.totalCapacity || 0,
    });
  }, [channelData?.length]);

  useEffect(() => {
    if (nodeConfigurations?.length === 0) {
      track("No Node Configured");
    } else {
      console.log("Node Configured");
      track("Node Configured");

      // check if all data is synced
      const allDataSynced = (nodeConfigurations || []).every((node) => node.status === 1);
      if (allDataSynced) {
        console.log("Node synced");
        track("All Data Synced", { nodeCount: nodeConfigurations?.length || 0 });
      }
    }
  }, [nodeConfigurationSuccess]);

  // check if all services are running
  useEffect(() => {
    // Register torq version
    if (servicesData?.version) {
      register({
        torqVersion: servicesData?.version,
      });
    }

    // Create an object of each lnd service typeString and statusString and register once
    if (servicesData?.lndServices?.length) {
      const lndServices = servicesData?.lndServices?.reduce((acc, service) => {
        acc[`lndService${service.typeString}`] = service.statusString;
        return acc;
      }, {} as { [key: string]: string });
      register(lndServices);
    }

    // Create an object of each lnd service typeString and torqService and register once
    if (servicesData?.torqServices?.length) {
      const torqServices = servicesData?.torqServices?.reduce((acc, service) => {
        acc[`torqService${service.typeString}`] = service.statusString;
        return acc;
      }, {} as { [key: string]: string });
      register(torqServices);
    }

    // Register main service status
    if (servicesData?.mainService) {
      register({
        torqServiceMainService: servicesData?.mainService?.statusString,
      });
    }
  }, [servicesData?.lndServices, servicesDataSuccess]);

  useEffect(() => {
    if (nodeQueryHasRun) {
      // Register each node public key and alias separately as a string array separated by comma.
      const nodePublicKeys = nodes?.map((node) => node.publicKey);
      const nodeAliases = nodes?.map((node) => node.alias);
      register({
        torqPublicKeys: nodePublicKeys?.join(", "),
        torqNodeAliases: nodeAliases?.join(", "),
        name: nodeAliases?.join(", "),
      });
    }
  }, [nodes?.length, nodeQueryHasRun]);

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
        <Outlet />
      </div>
    </div>
  );
}

export default DefaultLayout;
