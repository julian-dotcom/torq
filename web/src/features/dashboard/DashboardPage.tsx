import DashboardPageTemplate from "features/templates/dashboardPageTemplate/DashboardPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/dashboard/dashboard-page.module.scss";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "../templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import { useNavigate } from "react-router-dom";
import * as Routes from "constants/routes";
import { useLocation } from "react-router";
import {
  MoneyHand20Regular as TransactionIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Check20Regular as InvoiceIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
} from "@fluentui/react-icons";

import { NEW_ADDRESS, NEW_INVOICE, NEW_PAYMENT } from "constants/routes";
import SummaryCard from "components/summary/summaryCard/SummaryCard";
import SummaryNode from "components/summary/summaryNode/SummaryNode";
import { useGetChannelsQuery, useGetNodeConfigurationsQuery, useGetNodesWalletBalancesQuery } from "apiSlice";
import { channel } from "features/channels/channelsTypes";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import { nodeWalletBalances } from "../../apiTypes";

interface nodeSummary {
  nodeId: number;
  name: string;
  onChainBalance: number;
  localBalance: number;
  totalBalance: number;
  channels: number;
  capacity: number;

  walletBalances: nodeWalletBalances;
}

interface nodeChannelSummary {
  nodeId: number;
  localBalance: number;
  channels: number;
  capacity: number;
}

interface allNodesSummary {
  onChainBalance: number;
  localBalance: number;
  totalBalance: number;
  channels: number;
}

function CalculateTotals(nodeSummaries: nodeSummary[] | undefined): allNodesSummary {
  if (!nodeSummaries) return {} as allNodesSummary;

  return nodeSummaries.reduce(
    (acc, node) => {
      acc.localBalance += node.localBalance;
      acc.onChainBalance += node.onChainBalance;
      acc.channels += node.channels;
      acc.totalBalance += node.totalBalance;
      return acc;
    },
    { onChainBalance: 0, offChainBalance: 0, localBalance: 0, totalBalance: 0, channels: 0 }
  );
}
function DashboardPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const { data: nodesWalletBalances } = useGetNodesWalletBalancesQuery();
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !nodeConfigurations, pollingInterval: 10000 });

  type channelReduceReturnType = Record<number, nodeChannelSummary>;

  //process channels based on node
  const nodesOpenChannelSummary = channelsResponse?.data?.reduce<channelReduceReturnType>((acc, channel) => {
    let channelsSummary = acc[channel.nodeId];
    if (!channelsSummary) {
      channelsSummary = {
        channels: 0,
        onChainBalance: 0,
        localBalance: 0,
        totalBalance: 0,
        capacity: 0,
      } as nodeSummary;

      acc[channel.nodeId] = channelsSummary;
    }

    channelsSummary.channels += 1;
    channelsSummary.capacity += channel.capacity;
    channelsSummary.localBalance += channel.localBalance;
    return acc;
  }, {});

  //nodes
  const nodeSummaries = nodeConfigurations?.map((x) => {
    const node = {
      nodeId: x.nodeId,
      name: x.name,
    } as nodeSummary;

    const walletBalances = nodesWalletBalances?.find((x) => x.nodeId === node.nodeId);
    if (walletBalances) {
      node.walletBalances = walletBalances;
      node.onChainBalance = walletBalances.totalBalance;
    }

    const channelSummary = nodesOpenChannelSummary ? nodesOpenChannelSummary[node.nodeId] : undefined;

    if (channelSummary) {
      node.channels = channelSummary.channels;
      node.localBalance = channelSummary.localBalance;
      node.capacity = channelSummary.capacity;
      node.totalBalance = node.localBalance + node.onChainBalance;
    }

    return node;
  });

  const totalsSummary = CalculateTotals(nodeSummaries);

  const controls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<ChannelsIcon />}
            onClick={() => {
              mixpanel.track("Navigate to Open Channel");
              navigate(Routes.OPEN_CHANNEL, { state: { background: location } });
            }}
          >
            {t.openChannel}
          </Button>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<TransactionIcon />}
            onClick={() => {
              mixpanel.track("Navigate to New Payment");
              navigate(NEW_PAYMENT, { state: { background: location } });
            }}
          >
            {t.newPayment}
          </Button>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<InvoiceIcon />}
            onClick={() => {
              navigate(NEW_INVOICE, { state: { background: location } });
              mixpanel.track("Navigate to New Invoice");
            }}
          >
            {t.header.newInvoice}
          </Button>

          <Button
            buttonColor={ColorVariant.success}
            icon={<NewOnChainAddressIcon />}
            hideMobileText={true}
            onClick={() => {
              navigate(NEW_ADDRESS, { state: { background: location } });
              mixpanel.track("Navigate to New OnChain Address");
            }}
          >
            {t.newAddress}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  return (
    <DashboardPageTemplate title={t.dashboard} welcomeMessage={t.dashboardPage.welcome}>
      {controls}
      <div className={styles.dashboardWrapper}>
        <div className={styles.summaryCardContainer}>
          <SummaryCard
            heading={t.dashboardPage.totalBalance}
            value={totalsSummary?.totalBalance}
            valueLabel={t.dashboardPage.btc}
          ></SummaryCard>
          <SummaryCard
            heading={t.dashboardPage.totalOnChainBalance}
            value={totalsSummary?.onChainBalance}
            valueLabel={t.dashboardPage.btc}
          ></SummaryCard>
          <SummaryCard
            heading={t.dashboardPage.totalOffChainBalance}
            value={totalsSummary?.localBalance}
            valueLabel={t.dashboardPage.btc}
          ></SummaryCard>
          <SummaryCard
            heading={t.dashboardPage.totalChannelCount}
            value={totalsSummary?.channels}
            valueLabel={""}
          ></SummaryCard>
        </div>

        {nodeSummaries?.map((node) => {
          return (
            <SummaryNode nodeName={node.name ?? ""} key={`node-${node.nodeId}`}>
              <div className={styles.summaryNodeContainer}>
                <SummaryCard
                  heading={t.dashboardPage.totalOnChainBalance}
                  value={node.onChainBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCard}
                  details={
                    <div>
                      <dl>
                        <dt>{t.dashboardPage.onChain.confirmed}</dt>
                        <dd>{node?.walletBalances?.confirmedBalance}</dd>
                        <dt>{t.dashboardPage.onChain.unconfirmed}</dt>
                        <dd>{node?.walletBalances?.unconfirmedBalance}</dd>
                        <dt>{t.dashboardPage.onChain.lockedBalance}</dt>
                        <dd>{node?.walletBalances?.lockedBalance}</dd>
                      </dl>
                    </div>
                  }
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.totalOffChainBalance}
                  value={node.localBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCard}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.totalBalance}
                  value={node.totalBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCard}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.channels}
                  value={node.channels}
                  valueLabel={""}
                  summaryClassOverride={styles.nodeSummaryCard}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.capacity}
                  value={node.capacity}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCard}
                ></SummaryCard>
              </div>
            </SummaryNode>
          );
        })}
      </div>
    </DashboardPageTemplate>
  );
}

export default DashboardPage;
