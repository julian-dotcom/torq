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

interface nodeSummary {
  onChainBalance: number;
  offChainBalance: number;
  localBalance: number;
  totalBalance: number;
  channels: number;
  capacity: number;
}

interface allNodesSummary {
  onChainBalance: number;
  offChainBalance: number;
  localBalance: number;
  totalBalance: number;
  channels: number;
}

function CalculateTotals(nodesSummary: Record<string, nodeSummary>): allNodesSummary | undefined {
  if (!nodesSummary) return undefined;

  return Object.entries(nodesSummary).reduce(
    (acc, [_, nodeSummary]) => {
      acc.localBalance += nodeSummary.localBalance;
      acc.offChainBalance += nodeSummary.offChainBalance;
      acc.onChainBalance += nodeSummary.onChainBalance;
      acc.channels += nodeSummary.channels;
      acc.totalBalance += nodeSummary.totalBalance;
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

  type channelReduceReturnType = Record<number, nodeSummary>;

  //process channels based on node
  const nodesSummary = channelsResponse?.data?.reduce<channelReduceReturnType>((acc, channel) => {
    if (!acc[channel.nodeId])
      acc[channel.nodeId] = {
        channels: 0,
        onChainBalance: 0,
        offChainBalance: 0,
        localBalance: 0,
        totalBalance: 0,
        capacity: 0,
      } as nodeSummary;

    console.log(nodesWalletBalances);
    const onChainBalance = 0;
    const nodeSummary = acc[channel.nodeId];
    nodeSummary.channels += 1;
    nodeSummary.capacity += channel.capacity;
    nodeSummary.onChainBalance += onChainBalance; // will require interegation?
    nodeSummary.localBalance += channel.localBalance;
    nodeSummary.totalBalance += channel.localBalance + onChainBalance;
    return acc;
  }, {});

  const totalsSummary = CalculateTotals(nodesSummary);

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
            value={totalsSummary?.offChainBalance}
            valueLabel={t.dashboardPage.btc}
          ></SummaryCard>
          <SummaryCard
            heading={t.dashboardPage.totalChannelCount}
            value={totalsSummary?.channels}
            valueLabel={""}
          ></SummaryCard>
        </div>

        {nodesSummary &&
          nodeConfigurations?.map((node) => {
            const nodeSummary = nodesSummary[node?.nodeId];
            if (!nodeSummary) return <></>;

            return (
              <SummaryNode nodeName={node.name ?? ""} key={`node-${node.nodeId}`}>
                <div className={styles.summaryNodeContainer}>
                  <SummaryCard
                    heading={t.dashboardPage.totalOnChainBalance}
                    value={nodeSummary.onChainBalance}
                    valueLabel={t.dashboardPage.btc}
                    summaryClassOverride={styles.nodeSummaryCard}
                    details={
                      <div>
                        <dl>
                          <dt>Confirmed</dt>
                          <dd>600,512,313</dd>
                          <dt>Unconfirmed</dt>
                          <dd>100,123,543</dd>
                          <dt>Locked Balance</dt>
                          <dd>0</dd>
                        </dl>
                      </div>
                    }
                  ></SummaryCard>
                  <SummaryCard
                    heading={t.dashboardPage.totalOffChainBalance}
                    value={nodeSummary.offChainBalance}
                    valueLabel={t.dashboardPage.btc}
                    summaryClassOverride={styles.nodeSummaryCard}
                  ></SummaryCard>
                  <SummaryCard
                    heading={t.dashboardPage.totalBalance}
                    value={nodeSummary.totalBalance}
                    valueLabel={t.dashboardPage.btc}
                    summaryClassOverride={styles.nodeSummaryCard}
                  ></SummaryCard>
                  <SummaryCard
                    heading={t.dashboardPage.channels}
                    value={nodeSummary.channels}
                    valueLabel={""}
                    summaryClassOverride={styles.nodeSummaryCard}
                  ></SummaryCard>
                  <SummaryCard
                    heading={t.dashboardPage.capacity}
                    value={nodeSummary.capacity}
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
