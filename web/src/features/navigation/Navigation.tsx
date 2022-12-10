import { useAppDispatch } from "store/hooks";
import { toggleNav } from "./navSlice";
import classNames from "classnames";
import MenuItem from "./MenuItem";
import NavCategory from "./NavCategory";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import {
  Navigation20Regular as CollapseIcon,
  ArrowForward20Regular as ForwardsIcon,
  Autosum20Regular as SummaryIcon,
  MoneyHand20Regular as PaymentsIcon,
  KeyMultiple20Regular as OnChainTransactionIcon,
  Check20Regular as InvoicesIcon,
  LockClosed20Regular as LogoutIcon,
  Settings20Regular as SettingsIcon,
  ArrowRouting20Regular as ChannelsIcon,
  // Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import * as routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";

function Navigation() {
  const dispatch = useAppDispatch();
  const { t } = useTranslations();

  return (
    <div className={classNames(styles.navigation)}>
      <div className={styles.logoWrapper}>
        <div className={classNames(styles.logo)}>
          <TorqLogo />
        </div>

        {/*<div className={classNames(styles.eventsButton)}>*/}
        {/*  <EventsIcon />*/}
        {/*</div>*/}

        <div className={styles.collapseButton} id={"collapse-navigation"} onClick={() => dispatch(toggleNav())}>
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

        <NavCategory text={"Manage"} collapsed={false}>
          <MenuItem text={t.channels} icon={<ChannelsIcon />} routeTo={"/manage/channels"} />
          {/*<MenuItem text={"Tags"} icon={<TagsIcon />} routeTo={"/manage/tags"} />*/}
        </NavCategory>

        <NavCategory text={"Transactions"} collapsed={false}>
          <MenuItem text={"Payments"} icon={<PaymentsIcon />} routeTo={`/${routes.TRANSACTIONS}/${routes.PAYMENTS}`} />
          <MenuItem text={"Invoices"} icon={<InvoicesIcon />} routeTo={`/${routes.TRANSACTIONS}/${routes.INVOICES}`} />
          <MenuItem
            text={"On-Chain"}
            icon={<OnChainTransactionIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.ONCHAIN}`}
          />
        </NavCategory>
      </div>

      <div className={classNames(styles.bottomWrapper)}>
        <MenuItem text={t.settings} icon={<SettingsIcon />} routeTo={"/settings"} />
        <MenuItem text={t.logout} icon={<LogoutIcon />} routeTo={"/logout"} />
      </div>
    </div>
  );
}

export default Navigation;
