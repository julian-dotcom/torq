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
  Signature20Regular as MessageVerificationIcon,
  // Flash20Regular as WorkflowsIcon,
  Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import * as routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import NetworkSelector from "./NetworkSelector";
import { useEffect } from "react";
import { useGetSettingsQuery } from "apiSlice";
import MenuButtonItem from "./MenuButtonItem";

function Navigation() {
  const dispatch = useAppDispatch();
  const { data: settingsData } = useGetSettingsQuery();
  const { t } = useTranslations();
  const { boot } = useIntercom();

  useEffect(() => {
    boot({
      userId: settingsData?.torqUuid,
      customLauncherSelector: "#intercom-launcher",
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

        <NavCategory text={t.analyse} collapsed={false}>
          <MenuItem text={t.summary} icon={<SummaryIcon />} routeTo={"/"} />
          <MenuItem text={t.forwards} icon={<ForwardsIcon />} routeTo={"/analyse/forwards"} />
          {/*<MenuItem text={"Inspect"} icon={<InspectIcon />} routeTo={"/inspect"} />*/}
        </NavCategory>

        <NavCategory text={t.manage} collapsed={false}>
          <MenuItem text={t.channels} icon={<ChannelsIcon />} routeTo={"/manage/channels"} />
          {/*<MenuItem text={t.automation} icon={<WorkflowsIcon />} routeTo={"/manage/workflows"} />*/}
          <MenuItem text={t.tags} icon={<TagsIcon />} routeTo={"/manage/tags"} />
          <MenuItem
            text={t.MessageVerification}
            icon={<MessageVerificationIcon />}
            routeTo={routes.MESSAGE_VERIFICATION}
            withBackground={true}
          />
        </NavCategory>

        <NavCategory text={t.transactions} collapsed={false}>
          <MenuItem text={t.payments} icon={<PaymentsIcon />} routeTo={`/${routes.TRANSACTIONS}/${routes.PAYMENTS}`} />
          <MenuItem text={t.invoices} icon={<InvoicesIcon />} routeTo={`/${routes.TRANSACTIONS}/${routes.INVOICES}`} />
          <MenuItem
            text={t.onChain}
            icon={<OnChainTransactionIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.ONCHAIN}`}
          />
        </NavCategory>
      </div>

      <div className={classNames(styles.bottomWrapper)}>
        <MenuButtonItem text={t.helpAndBugsMenuItem} icon={<HelpIcon />} id={"intercom-launcher"} />
        <MenuItem text={t.settings} icon={<SettingsIcon />} routeTo={"/settings"} />
        <MenuItem text={t.logout} icon={<LogoutIcon />} routeTo={"/logout"} />
      </div>
    </div>
  );
}

export default Navigation;
