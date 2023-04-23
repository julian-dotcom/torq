import { useAppDispatch, useAppSelector } from "store/hooks";
import { userEvents } from "utils/userEvents";
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
  Molecule20Regular as PeersIcon,
} from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import * as routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import NetworkSelector from "./NetworkSelector";
import { useGetAutoLoginSettingQuery } from "apiSlice";
import MenuButtonItem from "./MenuButtonItem";

function Navigation() {
  const dispatch = useAppDispatch();
  const { data: autoLogin } = useGetAutoLoginSettingQuery();
  const { t } = useTranslations();
  const { track, register } = userEvents();

  const hidden = useAppSelector(selectHidden);

  function toggleNavHandler() {
    track("Toggle Navigation");
    register({ navigation_collapsed: !hidden });
    dispatch(toggleNav());
  }

  return (
    <div className={classNames(styles.navigation)}>
      <div className={styles.logoWrapper}>
        <div className={classNames(styles.logo)}>
          <TorqLogo />
        </div>

        <NetworkSelector />

        <div
          className={styles.collapseButton}
          id={"collapse-navigation"}
          onClick={toggleNavHandler}
          data-intercom-target="collapse-nav-button"
        >
          <CollapseIcon />
        </div>
      </div>

      <div className={styles.mainNavWrapper}>
        <NavCategory text={t.dashboard} collapsed={false} intercomTarget={"dashboard-nav-section"}>
          <MenuItem
            intercomTarget="dashboard-nav-button"
            text={t.dashboard}
            icon={<DashboardIcon />}
            routeTo={"/"}
            onClick={() => {
              track("Navigate to Dashboard");
            }}
          />
        </NavCategory>
        <NavCategory text={t.analyse} collapsed={false} intercomTarget={"forwards-nav-section"}>
          <MenuItem
            intercomTarget="forwards-summary-nav-button"
            text={t.summary}
            icon={<SummaryIcon />}
            routeTo={`/${routes.ANALYSE}/${routes.FORWARDS_SUMMARY}`}
            onClick={() => {
              track("Navigate to Summary");
            }}
          />

          <MenuItem
            intercomTarget="forwards-nav-button"
            text={t.forwards}
            icon={<ForwardsIcon />}
            routeTo={"/analyse/forwards"}
            onClick={() => {
              track("Navigate to Forwards");
            }}
          />
        </NavCategory>

        <NavCategory text={t.channels} collapsed={false} intercomTarget={"channels-nav-section"}>
          <MenuItem
            intercomTarget="open-channels-nav-button"
            text={t.openChannels}
            icon={<ChannelsIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.OPEN_CHANNELS}`}
            onClick={() => {
              track("Navigate to Open Channels");
            }}
          />
          <MenuItem
            intercomTarget="pending-channels-nav-button"
            text={t.pendingChannels}
            icon={<ChannelsPendingIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.PENDING_CHANNELS}`}
            onClick={() => {
              track("Navigate to Pending Channels");
            }}
          />
          <MenuItem
            intercomTarget="closed-channels-nav-button"
            text={t.closedChannels}
            icon={<ChannelsClosedIcon />}
            routeTo={`/${routes.CHANNELS}/${routes.CLOSED_CHANNELS}`}
            onClick={() => {
              track("Navigate to Closed Channels");
            }}
          />
        </NavCategory>

        <NavCategory text={t.manage} collapsed={false} intercomTarget={"manage-nav-section"}>
          <MenuItem
            intercomTarget="automation-nav-button"
            text={t.automation}
            icon={<WorkflowsIcon />}
            routeTo={"/manage/workflows"}
            onClick={() => {
              track("Navigate to Workflows");
            }}
          />
          <MenuItem
            intercomTarget="tags-nav-button"
            text={t.tags}
            icon={<TagsIcon />}
            routeTo={"/manage/tags"}
            onClick={() => {
              track("Navigate to Tags");
            }}
          />
          <MenuItem
            intercomTarget="peers-nav-button"
            text={t.peers}
            icon={<PeersIcon />}
            routeTo={`/manage/${routes.PEERS}`}
            onClick={() => {
              track("Navigate to Peers");
            }}
          />

          <MenuItem
            intercomTarget="messages-nav-button"
            text={t.MessageVerification}
            icon={<MessageVerificationIcon />}
            routeTo={routes.MESSAGE_VERIFICATION}
            withBackground={true}
            onClick={() => {
              track("Navigate to Message Verification");
            }}
          />
        </NavCategory>

        <NavCategory text={t.transactions} collapsed={false} intercomTarget={"transactions-nav-section"}>
          <MenuItem
            intercomTarget="payments-nav-button"
            text={t.payments}
            icon={<PaymentsIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.PAYMENTS}`}
            onClick={() => {
              track("Navigate to Payments");
            }}
          />
          <MenuItem
            intercomTarget="invoices-nav-button"
            text={t.invoices}
            icon={<InvoicesIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.INVOICES}`}
            onClick={() => {
              track("Navigate to Invoices");
            }}
          />
          <MenuItem
            intercomTarget="on-chain-nav-button"
            text={t.onChain}
            icon={<OnChainTransactionIcon />}
            routeTo={`/${routes.TRANSACTIONS}/${routes.ONCHAIN}`}
            onClick={() => {
              track("Navigate to OnChain Transactions");
            }}
          />
        </NavCategory>
      </div>

      <div className={classNames(styles.bottomWrapper)} data-intercom-target={"bottom-nav-section"}>
        <MenuButtonItem
          intercomTarget="intercom-launcher"
          text={t.helpAndBugsMenuItem}
          icon={<HelpIcon />}
          id={"intercom-launcher"}
          onClick={() => {
            track("Toggle Help");
          }}
        />
        <MenuItem
          intercomTarget="settings-nav-button"
          text={t.settings}
          icon={<SettingsIcon />}
          routeTo={"/settings"}
          onClick={() => {
            track("Navigate to Settings");
          }}
        />
        {!autoLogin && (
          <MenuItem
            intercomTarget="logout-nav-button"
            text={t.logout}
            icon={<LogoutIcon />}
            routeTo={"/logout"}
            onClick={() => {
              track("Logout");
            }}
          />
        )}
      </div>
    </div>
  );
}

export default Navigation;
