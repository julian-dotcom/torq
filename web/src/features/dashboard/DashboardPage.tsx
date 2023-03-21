import DashboardPageTemplate from "features/templates/dashboardPageTemplate/DashboardPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/dashboard/dashboard-page.module.scss";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "../templates/tablePageTemplate/TablePageTemplate";
import Button, { ButtonPosition, ColorVariant, SizeVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import { useNavigate } from "react-router-dom";
import * as Routes from "constants/routes";
import { NEW_ADDRESS, NEW_INVOICE, NEW_PAYMENT } from "constants/routes";
import { useLocation } from "react-router";
import {
  ArrowRouting20Regular as ChannelsIcon,
  Check20Regular as InvoiceIcon,
  Checkmark20Regular as CheckmarkIcon,
  Copy16Regular as CopyIcon,
  Dismiss20Regular as DismissIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
  MoneyHand20Regular as TransactionIcon,
  Question20Regular as QuestionIcon,
} from "@fluentui/react-icons";
import { format } from "d3";
import SummaryCard from "components/summary/summaryCard/SummaryCard";
import SummaryNode from "components/summary/summaryNode/SummaryNode";
import {
  useGetChannelsPendingQuery,
  useGetChannelsQuery,
  useGetNodesInformationByCategoryQuery,
  useGetNodesWalletBalancesQuery,
} from "apiSlice";
import { channel } from "features/channels/channelsTypes";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import { nodeAddress, nodeStatus, nodeWalletBalances } from "apiTypes";
import { ChannelPending } from "../channelsPending/channelsPendingTypes";
import classNames from "classnames";
import React, { useEffect, useState } from "react";
import ToastContext from "../toast/context";
import { copyToClipboard } from "../../utils/copyToClipboard";
import { toastCategory } from "../toast/Toasts";
import Modal from "../modal/Modal";

const tformtatter = format(",.0f");

interface nodeSummary {
  nodeId: number;
  name: string;
  onChainBalance: number;
  localBalance: number;
  totalBalance: number;
  channels: number;
  capacity: number;
  closingChannels: number;
  openingChannels: number;
  walletBalances: nodeWalletBalances;
  status: number;
  addresses: nodeAddress[];

  publicKey: string;
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

  return nodeSummaries
    .filter((x) => x.status === nodeStatus.active)
    .reduce(
      (acc, node) => {
        acc.localBalance += node.localBalance ?? 0;
        acc.onChainBalance += node.onChainBalance ?? 0;
        acc.channels += node.channels ?? 0;
        acc.totalBalance += node.totalBalance ?? 0;
        return acc;
      },
      { onChainBalance: 0, offChainBalance: 0, localBalance: 0, totalBalance: 0, channels: 0 }
    );
}
function DashboardPage() {
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const navigate = useNavigate();
  const location = useLocation();
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const { data: nodesWalletBalances } = useGetNodesWalletBalancesQuery(activeNetwork);
  const { data: nodes, isSuccess: nodeQueryHasRun } = useGetNodesInformationByCategoryQuery(activeNetwork);

  const [showNodeSetupModal, setShowNodeSetupModal] = useState(false);

  useEffect(() => {
    if (nodeQueryHasRun && (!nodes || nodes.length === 0)) {
      setShowNodeSetupModal(true);
    }
  }, [nodeQueryHasRun]);

  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !nodes, pollingInterval: 10000 });

  const pendingChannelsResponse = useGetChannelsPendingQuery<{
    data: Array<ChannelPending>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork });

  const copyText = async (copyText: string) => {
    await copyToClipboard(copyText || "");
    toastRef?.current?.addToast(`${copyText} copied to clipboard`, toastCategory.success);
  };

  const handleConfirmationModalClose = () => {
    mixpanel.track("No Local Node", {
      source: location?.pathname,
      "Go to settings": false,
    });
    setShowNodeSetupModal(false);
  };

  const handleModalSettingsClick = () => {
    setShowNodeSetupModal(false);
    mixpanel.track("No Local Node", {
      source: location?.pathname,
      "Go to settings": true,
    });
    navigate("/settings", { replace: true });
  };

  type channelReduceReturnType = Record<number, nodeChannelSummary>;

  //process channels based on node
  const nodesOpenChannelSummary = channelsResponse?.data?.reduce<channelReduceReturnType>((acc, channel) => {
    let channelsSummary = acc[channel.nodeId];
    if (!channelsSummary) {
      channelsSummary = {
        nodeId: channel.nodeId,
        channels: 0,
        onChainBalance: 0,
        localBalance: 0,
        totalBalance: 0,
        capacity: 0,
      } as nodeChannelSummary;

      acc[channel.nodeId] = channelsSummary;
    }

    channelsSummary.channels += 1;
    channelsSummary.capacity += channel.capacity;
    channelsSummary.localBalance += channel.localBalance;
    return acc;
  }, {});

  //nodes
  const nodeSummaries = nodes?.map((x) => {
    const node = {
      nodeId: x.nodeId,
      name: x.alias,
      status: x.status,
      addresses: x.addresses,
      publicKey: x.publicKey,
      openingChannels: 0,
      closingChannels: 0,
    } as nodeSummary;

    const walletBalances =
      nodesWalletBalances?.find((x) => x.nodeId === node.nodeId) ??
      ({
        nodeId: 0,
        totalBalance: 0,
        confirmedBalance: 0,
        unconfirmedBalance: 0,
        lockedBalance: 0,
        reservedBalanceAnchorChan: 0,
      } as nodeWalletBalances);

    node.walletBalances = walletBalances;
    node.onChainBalance = walletBalances.totalBalance;

    const channelSummary = nodesOpenChannelSummary ? nodesOpenChannelSummary[node.nodeId] : undefined;

    if (channelSummary) {
      node.channels = channelSummary.channels;
      node.localBalance = channelSummary.localBalance;
      node.capacity = channelSummary.capacity;
      node.totalBalance = node.localBalance + node.onChainBalance;
    }

    const pendingChannels = pendingChannelsResponse?.data?.filter((x) => x.nodeId === node.nodeId);

    pendingChannels?.forEach((pendingChan) => {
      switch (pendingChan.status.toLowerCase()) {
        case "closing":
          node.closingChannels += 1;
          break;
        case "opening":
          node.openingChannels += 1;
      }
    });

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
            <SummaryNode nodeName={node.name ?? ""} key={`node-${node.nodeId}`} status={node.status}>
              <div className={styles.summaryNodeContainer}>
                <SummaryCard
                  heading={t.dashboardPage.totalOnChainBalance}
                  value={node.onChainBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                  details={
                    <div className={styles.nodeSummaryDetailsContainer}>
                      <div
                        className={classNames(styles.nodeSummaryDetailsColumn, styles.nodeSummaryDetailsTitleColumn)}
                      >
                        <div>{t.dashboardPage.onChainDetail.confirmed}</div>
                        <div>{t.dashboardPage.onChainDetail.unconfirmed}</div>
                        <div>{t.dashboardPage.onChainDetail.lockedBalance}</div>
                      </div>
                      <div
                        className={classNames(styles.nodeSummaryDetailsColumn, styles.nodeSummaryDetailsNumberColumn)}
                      >
                        <div>{tformtatter(node?.walletBalances?.confirmedBalance)}</div>
                        <div>{tformtatter(node?.walletBalances?.unconfirmedBalance)}</div>
                        <div>{tformtatter(node?.walletBalances?.lockedBalance)}</div>
                      </div>
                    </div>
                  }
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.totalOffChainBalance}
                  value={node.localBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.totalBalance}
                  value={node.totalBalance}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.channels}
                  value={node.channels}
                  valueLabel={""}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                  details={
                    <div className={styles.nodeSummaryDetailsContainer}>
                      <div
                        className={classNames(styles.nodeSummaryDetailsColumn, styles.nodeSummaryDetailsTitleColumn)}
                      >
                        <div>{t.dashboardPage.channelsDetail.opening}</div>
                        <div>{t.dashboardPage.channelsDetail.closing}</div>
                      </div>
                      <div
                        className={classNames(styles.nodeSummaryDetailsColumn, styles.nodeSummaryDetailsNumberColumn)}
                      >
                        <div>{node?.openingChannels}</div>
                        <div>{node?.closingChannels}</div>
                      </div>
                    </div>
                  }
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.capacity}
                  value={node.capacity}
                  valueLabel={t.dashboardPage.btc}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                ></SummaryCard>
                <SummaryCard
                  heading={t.dashboardPage.publicKey}
                  value={node.channels}
                  valueLabel={""}
                  summaryClassOverride={styles.nodeSummaryCardOverride}
                  details={
                    <div className={styles.addressDetailsContainer}>
                      <div className={styles.addressDetailItem} title={node.publicKey}>
                        {node.publicKey}
                      </div>
                      <Button
                        onClick={() => copyText(node.publicKey)}
                        icon={<CopyIcon />}
                        buttonSize={SizeVariant.tiny}
                        buttonColor={ColorVariant.success}
                      />
                      {node.addresses &&
                        node.addresses.length > 0 &&
                        node.addresses.map((address, i) => {
                          return (
                            <div className={styles.addressDetailItemRow} key={"address-row" + i}>
                              <div className={styles.addressDetailItem} title={address.addr}>
                                {address.addr}
                              </div>
                              <Button
                                onClick={() => copyText(address.addr)}
                                icon={<CopyIcon />}
                                buttonSize={SizeVariant.tiny}
                                buttonColor={ColorVariant.success}
                              />
                            </div>
                          );
                        })}
                    </div>
                  }
                >
                  <div className={styles.addressContent} title={node.publicKey}>
                    {node.publicKey}
                  </div>
                </SummaryCard>
              </div>
            </SummaryNode>
          );
        })}
      </div>
      <Modal
        title={"Local Node not found"}
        icon={<QuestionIcon className={styles.settingsConfirmationIcon} />}
        onClose={handleConfirmationModalClose}
        show={showNodeSetupModal}
      >
        <div className={styles.deleteConfirm}>
          <p>
            It seems that you didn&apos;t configure a node. In order to use Torq, you will have to configure at least
            one Node.
          </p>
          <p>Do you want to go to the Settings page and add your Node now?</p>

          <div className={styles.settingsConfirmation}>
            <Button
              buttonColor={ColorVariant.success}
              buttonPosition={ButtonPosition.fullWidth}
              icon={<CheckmarkIcon />}
              onClick={handleModalSettingsClick}
              className={styles.settingsConfirmationButton}
            >
              {"Yes"}
            </Button>
            <Button
              buttonColor={ColorVariant.error}
              buttonPosition={ButtonPosition.fullWidth}
              icon={<DismissIcon />}
              onClick={handleConfirmationModalClose}
              className={styles.settingsConfirmationButton}
              id="no-settings-confirmation"
            >
              {"No"}
            </Button>
          </div>
        </div>
      </Modal>
    </DashboardPageTemplate>
  );
}

export default DashboardPage;
