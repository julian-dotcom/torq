import { useAppDispatch, useAppSelector } from "store/hooks";
import mixpanel from "mixpanel-browser";
import { useIntercom } from "react-use-intercom";
import { selectHidden, toggleNav } from "./navSlice";
import classNames from "classnames";
import MenuItem from "./MenuItem";
import NavCategory from "./NavCategory";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import {
  ChatHelp20Regular as HelpIcon,
  Navigation20Regular as CollapseIcon,
  ArrowForward20Regular as ForwardsIcon,
  Autosum20Regular as SummaryIcon,
  MoneyHand20Regular as PaymentsIcon,
  KeyMultiple20Regular as OnChainTransactionIcon,
  Check20Regular as InvoicesIcon,
  LockClosed20Regular as LogoutIcon,
  Settings20Regular as SettingsIcon,
  ArrowRouting20Regular as ChannelsIcon,
  ArrowWrapOff20Regular as ChannelsClosedIcon,
  ArrowRoutingRectangleMultiple20Regular as ChannelsPendingIcon,
  Signature20Regular as MessageVerificationIcon,
  Flash20Regular as WorkflowsIcon,
  Tag20Regular as TagsIcon,
  PanelSeparateWindow20Regular as DashboardIcon,
} from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import * as routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import NetworkSelector from "./NetworkSelector";
import { useEffect } from "react";
import { useGetAutoLoginSettingQuery, useGetSettingsQuery } from "apiSlice";
import MenuButtonItem from "./MenuButtonItem";

function Navigation() {
  const dispatch = useAppDispatch();
  const { data: settingsData } = useGetSettingsQuery();
  const { data: autoLogin } = useGetAutoLoginSettingQuery();
  const { t } = useTranslations();
  const { boot } = useIntercom();

  useEffect(() => {
    boot({
      userId: settingsData?.torqUuid,
      customLauncherSelector: "#intercom-launcher",
      hideDefaultLauncher: true,
    });
  }, [settingsData?.torqUuid]);

  const hidden = useAppSelector(selectHidden);

  function toggleNavHandler() {
    mixpanel.track("Toggle Navigation");
    mixpanel.register({ navigation_collapsed: !hidden });
    dispatch(toggleNav());
  }

  return (
    <div className={classNames(styles.navigation)}>
      <div className={styles.logoWrapper}>
        <div className={classNames(styles.logo)}>
          <TorqLogo />
        </div>

        <NetworkSelector />

        <div className={styles.collapseButton} id={"collapse-navigation"} onClick={toggleNavHandler}>
          <CollapseIcon />
        </div>
      </div>

      <div className={styles.mainNavWrapper}>
        {/*<MenuItem text={"Dashboard"} icon={<DashboardIcon />} routeTo={"/sadfa"} />*/}

        <NavCategory text={t.dashboard} collapsed={false}>
          <MenuItem
            text={t.dashboard}
            icon={<DashboardIcon />}
            routeTo={"/"}
            onClick={() => mixpanel.track("Navigate to Dashboard")}
          />
        </NavCategory>
        <NavCategory text={t.analyse} collapsed={false}>
          <MenuItem
            text={t.summary}
            icon={<SummaryIcon />}
            routeTo={"/"}
            onClick={() => mixpanel.track("Navigate to Summary")}
          />

          <MenuItem
            text={t.forwards}
            icon={<ForwardsIcon />}
            routeTo={"/analyse/forwards"}
            onClick={() => mixpanel.track("Navigate to Forwards")}
          />
        </NavCategory>

        <NavCategory text={t.channels} collapsed={false}>
          <MenuItem
            text={t.openChannels}
            icon={<ChannelsIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.OPEN_CHANNELS}`}
            onClick={() => mixpanel.track("Navigate to Open Channels")}
          />
          <MenuItem
            text={t.pendingChannels}
            icon={<ChannelsPendingIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.PENDING_CHANNELS}`}
            onClick={() => mixpanel.track("Navigate to Pending Channels")}
          />
          <MenuItem
            text={t.closedChannels}
            icon={<ChannelsClosedIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.CLOSED_CHANNELS}`}
            onClick={() => mixpanel.track("Navigate to Closed Channels")}
          />
        </NavCategory>

        <NavCategory text={t.manage} collapsed={false}>
          <MenuItem
            text={t.automation}
            icon={<WorkflowsIcon />}
            routeTo={"/manage/workflows"}
            onClick={() => mixpanel.track("Navigate to Workflows")}
          />
          <MenuItem
            text={t.tags}
            icon={<TagsIcon />}
            routeTo={"/manage/tags"}
            onClick={() => mixpanel.track("Navigate to Tags")}
          />

          <MenuItem
            text={t.MessageVerification}
            icon={<MessageVerificationIcon />}
            routeTo={routes.MESSAGE_VERIFICATION}
            withBackground={true}
            onClick={() => mixpanel.track("Navigate to Message Verification")}
          />
        </NavCategory>

        <NavCategory text={t.transactions} collapsed={false}>
          <MenuItem
            text={t.payments}
            icon={<PaymentsIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.PAYMENTS}`}
            onClick={() => mixpanel.track("Navigate to Payments")}
          />
          <MenuItem
            text={t.invoices}
            icon={<InvoicesIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.INVOICES}`}
            onClick={() => mixpanel.track("Navigate to Invoices")}
          />
          <MenuItem
            text={t.onChain}
            icon={<OnChainTransactionIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.ONCHAIN}`}
            onClick={() => mixpanel.track("Navigate to OnChain Transactions")}
          />
        </NavCategory>
      </div>

      <div className={classNames(styles.bottomWrapper)}>
        <MenuButtonItem
          text={t.helpAndBugsMenuItem}
          icon={<HelpIcon />}
          id={"intercom-launcher"}
          onClick={() => mixpanel.track("Toggle Help")}
        />
        <MenuItem
          text={t.settings}
          icon={<SettingsIcon />}
          routeTo={"/settings"}
          onClick={() => mixpanel.track("Navigate to Settings")}
        />
        {!autoLogin && (
          <MenuItem
            text={t.logout}
            icon={<LogoutIcon />}
            routeTo={"/logout"}
            onClick={() => mixpanel.track("Logout")}
          />
        )}
      </div>
    </div>
  );
}

export default Navigation;
