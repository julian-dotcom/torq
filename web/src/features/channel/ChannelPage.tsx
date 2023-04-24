import {
  ArrowRoutingRegular as ChannelsIcon,
  Delete16Regular as CloseIcon,
  Edit16Regular as UpdateIcon,
  Flag16Regular as EventFlagIcon,
  MoleculeRegular as NodeIcon,
} from "@fluentui/react-icons";
import {
  useGetChannelBalanceQuery,
  useGetChannelEventQuery,
  useGetChannelHistoryQuery,
  useGetChannelOnChainCostQuery,
  useGetChannelRebalancingQuery,
  useGetChannelsQuery,
  useGetFlowQuery,
} from "apiSlice";
import { Channel, ChannelEventResponse, ChannelHistoryResponse } from "features/channel/channelTypes";
import type { GetChannelHistoryData, GetFlowQueryParams } from "types/api";
import classNames from "classnames";
import * as d3 from "d3";
import { addDays, format } from "date-fns";
import React from "react";
import { useLocation, useNavigate, useParams } from "react-router";
import { useAppDispatch, useAppSelector } from "store/hooks";
import Button, { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import EventsCard from "features/eventsCard/EventsCard";
import { InputColorVaraint, InputSizeVariant, Select, Switch } from "components/forms/forms";
import Popover from "features/popover/Popover";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import BalanceChart from "./balanceChart/BalanceChart";
import styles from "./channel-page.module.scss";
import {
  selectBalanceChanID,
  selectEventChartKey,
  selectFlowKeys,
  selectProfitChartKey,
  updateBalanceChanID,
  updateEventChartKey,
  updateFlowKey,
  updateProfitChartKey,
} from "./channelSlice";
import EventsChart from "./eventsChart/EventsChart";
import FlowChart from "./flowChart/FlowChart";
import ProfitsChart from "./revenueChart/ProfitsChart";
import { selectActiveNetwork } from "features/network/networkSlice";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import PageTitle from "features/templates/PageTitle";
import useTranslations from "services/i18n/useTranslations";
import Tag, { TagColor } from "components/tags/Tag";
import { CLOSE_CHANNEL, UPDATE_CHANNEL } from "constants/routes";
import { userEvents } from "utils/userEvents";

const ft = d3.format(",.0f");

const eventNames = new Map([
  ["feeRate", "Fee rate"],
  ["baseFee", "Base fee"],
  ["minHtlc", "Min htlc"],
  ["maxHtlc", "Max htlc"],
  ["enabled", "Enabled"],
  ["disabled", "Disabled"],
]);

type ChannelPageProps = {
  chanId?: string;
};

function ChannelPage(_: ChannelPageProps) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const navigate = useNavigate();
  const location = useLocation();
  const currentPeriod = useAppSelector(selectTimeInterval);
  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [allToggle, setAllToggle] = React.useState(true);
  const [selectedEvents, setSelectedEvents] = React.useState({
    feeRate: true,
    baseFee: true,
    minHtlc: true,
    maxHtlc: true,
    enabled: true,
    disabled: true,
  });

  const { chanId } = useParams();

  const { data: channelDetails } = useGetChannelsQuery({ network: activeNetwork });

  const channelDetail = channelDetails?.find((c) => c.channelId === parseInt(chanId || ""));

  const flowQueryParams: GetFlowQueryParams = {
    from: from,
    to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
    chanIds: chanId || "all",
    network: activeNetwork,
  };
  const { data, isLoading } = useGetFlowQuery(flowQueryParams);

  const channelHistoryQueryData: GetChannelHistoryData = {
    params: { chanId: chanId || "all" },
    queryParams: {
      from: from,
      to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
      network: activeNetwork,
    },
  };

  const { data: balance } = useGetChannelBalanceQuery(channelHistoryQueryData);

  const { data: onChainCost } = useGetChannelOnChainCostQuery(channelHistoryQueryData);

  const { data: history } = useGetChannelHistoryQuery(channelHistoryQueryData);

  const { data: rebalancing } = useGetChannelRebalancingQuery(channelHistoryQueryData);

  const { data: event } = useGetChannelEventQuery(channelHistoryQueryData);

  const flowKey = useAppSelector(selectFlowKeys);
  const profitKey = useAppSelector(selectProfitChartKey);
  const eventKey = useAppSelector(selectEventChartKey);
  let balanceChanId = useAppSelector(selectBalanceChanID);

  if (balanceChanId.label === "") {
    balanceChanId = { value: 0, label: balance?.channelBalances ? balance.channelBalances[0]?.channelId : "" };
  }

  let totalCapacity = 0;
  if (history?.channels) {
    totalCapacity = history.channels
      .map((d: { capacity: number }) => {
        return d.capacity;
      })
      .reduce((partialSum: number, a: number) => partialSum + a, 0);
  }
  const profit: number =
    history?.revenueOut && onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? history?.revenueOut - onChainCost?.onChainCost - rebalancing?.rebalancingCost / 1000
      : 0;

  const totalCost: number =
    onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? onChainCost?.onChainCost + rebalancing?.rebalancingCost / 1000
      : 0;

  const selectedEventsCount = Array.from(Object.values(selectedEvents)).filter((d) => d).length;
  const historyAmountIn = history?.amountIn || 0;
  const historyAmountOut = history?.amountOut || 0;
  const historyAmountTotal = history?.amountTotal || 0;
  const historyRevenueOut = history?.revenueOut || 0;
  const historyRevenueIn = history?.revenueIn || 0;
  const historyRevenueTotal = history?.revenueTotal || 0;
  const historyCountIn = history?.countIn || 0;
  const historyCountOut = history?.countOut || 0;
  const historyCountTotal = history?.countTotal || 0;
  const historyRevenueOutAmountOut =
    history?.revenueOut && history?.amountOut ? (history?.revenueOut / history?.amountOut) * 1000 * 1000 : 0;
  const historyRevenueInAmountIn =
    history?.revenueIn && history?.amountIn ? (history?.revenueIn / history?.amountIn) * 1000 * 1000 : 0;
  const historyRevenueTotalAmountTotal =
    history?.revenueTotal && history?.amountTotal ? (history?.revenueTotal / history?.amountTotal) * 1000 * 1000 : 0;
  const rebalancingCostBy1000 = rebalancing?.rebalancingCost ? rebalancing?.rebalancingCost / 1000 : 0;
  const onchainPlusRebalancingCostBy1000 =
    onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? onChainCost?.onChainCost + rebalancing?.rebalancingCost / 1000
      : 0;
  const onchainCost = onChainCost?.onChainCost || 0;
  const historyReveueOutMinusCost = history?.revenueOut ? (history?.revenueOut - totalCost) / history?.revenueOut : 0;
  const historyAmountTotalCapacity = history?.amountTotal ? history?.amountTotal / totalCapacity : 0;
  const historyAmountInOut =
    history?.amountIn && history?.amountOut
      ? Math.min(history?.amountIn, history?.amountOut) / Math.max(history?.amountIn, history?.amountOut)
      : 0;
  const historyCountInOut =
    history?.countIn && history?.countOut
      ? Math.min(history?.countIn, history?.countOut) / Math.max(history?.countIn, history?.countOut)
      : 0;
  const historyCapacity = history?.channels?.length ? history?.channels[balanceChanId.value].capacity : 0;

  const breadcrumbs =
    (!isLoading &&
      history &&
      (history.channels || [])
        .map((d: Channel, _: number) => {
          return d.shortChannelId;
        })
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        .filter((value: any, index: number, self: any[]) => {
          return self.indexOf(value) === index;
        })
        .join(", ")) ||
    "";
  let channelBalanceOptions = [{ value: 0, label: "" }];
  if (balance?.channelBalances) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    channelBalanceOptions = balance.channelBalances.map((d: any, i: number) => {
      return { value: i, label: d.channelId };
    });
  }

  const isSingleChannel = channelDetails?.length && history?.channels?.length === 1;
  return (
    <PopoutPageTemplate
      title={t.inspectChannel}
      show={true}
      onClose={() => {
        const loc = location?.state?.background?.pathname || "/";
        navigate(loc);
      }}
      fullWidth={true}
    >
      <PageTitle
        title={
          (channelDetails || [])
            .filter((chan) => (chanId || "").split(",").includes(chan.channelId.toString()))
            .map((chan) => chan.peerAlias)
            // Keep only unique names
            .filter((value, index, self) => {
              return self.indexOf(value) === index;
            })
            .join(" | ") || breadcrumbs
        }
        className={styles.detailsPageTitle}
        breadcrumbs={[breadcrumbs || ""]}
      >
        <TimeIntervalSelect />
      </PageTitle>

      <div className={styles.channelWrapper}>
        <div className={styles.actionRow}>
          <div className={styles.actionButtons}>
            {isSingleChannel && (
              <>
                <LinkButton
                  intercomTarget={"inspect-update-channel"}
                  to={`${UPDATE_CHANNEL}?nodeId=${channelDetail?.nodeId}&channelId=${channelDetail?.channelId}`}
                  state={{ background: location.state.background }}
                  hideMobileText={true}
                  icon={<UpdateIcon />}
                  buttonColor={ColorVariant.success}
                  buttonSize={SizeVariant.small}
                  onClick={() => {
                    track("Navigate to Update Channel", {
                      nodeId: channelDetail?.nodeId,
                      channelId: channelDetail?.channelId,
                    });
                  }}
                >
                  {t.update}
                </LinkButton>

                <LinkButton
                  intercomTarget={"inspect-close-channel"}
                  to={`${CLOSE_CHANNEL}?nodeId=${channelDetail?.nodeId}&channelId=${channelDetail?.channelId}`}
                  state={{ background: location.state.background }}
                  hideMobileText={true}
                  icon={<CloseIcon />}
                  buttonSize={SizeVariant.small}
                  buttonColor={ColorVariant.error}
                  onClick={() => {
                    track("Navigate to Close Channel", {
                      nodeId: channelDetail?.nodeId,
                      channelId: channelDetail?.channelId,
                    });
                  }}
                >
                  {t.close}
                </LinkButton>
              </>
            )}
          </div>
          <div className={styles.tags} data-intercom-target={"inspect-channel-tag-list"}>
            {((history?.channels && history.channels[0].tags) || []).map((tag) => (
              <Tag key={"tag-" + tag.tagId} label={tag.name} colorVariant={TagColor[tag.style]} />
            ))}
            {isSingleChannel && (
              <>
                <LinkButton
                  intercomTarget={"inspect-tag-channel"}
                  to={`/tag-channel/${chanId}`}
                  state={{ background: location.state.background }}
                  onClick={() => {
                    track("Navigate to Tag Channel", {
                      channelId: chanId,
                    });
                  }}
                  icon={<ChannelsIcon />}
                  buttonSize={SizeVariant.small}
                  buttonColor={ColorVariant.disabled}
                />
                <LinkButton
                  intercomTarget={"inspect-tag-node"}
                  to={`/tag-node/${channelDetail?.peerNodeId}`}
                  state={{ background: location.state.background }}
                  onClick={() => {
                    track("Navigate to Tag Node", {
                      nodeId: channelDetail?.peerNodeId,
                    });
                  }}
                  icon={<NodeIcon />}
                  buttonSize={SizeVariant.small}
                  buttonColor={ColorVariant.disabled}
                />
              </>
            )}
          </div>
        </div>

        <div
          className={classNames(styles.pageRow, styles.channelSummary)}
          data-intercom-target={"inspect-channel-summary"}
        >
          <div className={styles.shortColumn}>
            <div className={styles.card} data-intercom-target={"inspect-channel-revenue-card"}>
              <div className={styles.heading}>{t.channelPage.revenue.header}</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.revenue.forwardingFees}</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.revenue.channelLeases}</div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.revenue.total}</div>
                <div className={styles.rowValue}>{history?.revenueOut}</div>
              </div>
            </div>
            <div className={styles.card} data-intercom-target={"inspect-channel-expenses-card"}>
              <div className={styles.heading}>{t.channelPage.expenses.heading}</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.expenses.rebalancing}</div>
                <div className={classNames(styles.rowValue)}>{ft(rebalancingCostBy1000)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.expenses.openAndClose}</div>
                <div className={classNames(styles.rowValue)}>{ft(onchainCost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.expenses.total}</div>
                <div className={classNames(styles.rowValue)}>{ft(onchainPlusRebalancingCostBy1000)}</div>
              </div>
            </div>
            <div className={styles.card} data-intercom-target={"inspect-channel-profit-card"}>
              <div className={styles.heading}>{t.channelPage.profit.heading}</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.profit.total}</div>
                <div className={classNames(styles.rowValue)}>{ft(profit)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.profit.grossProfitMargin}</div>
                <div className={classNames(styles.rowValue)}>{d3.format(".2%")(historyReveueOutMinusCost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.profit.turnover}</div>
                <div className={classNames(styles.rowValue)}>{d3.format(",.2")(historyAmountTotalCapacity)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Return on Capacity</div>
                <div className={classNames(styles.rowValue)}>{d3.format(".2%")(profit / totalCapacity)}</div>
              </div>
            </div>

            <div
              className={classNames(styles.card, styles.currentCard)}
              data-intercom-target={"inspect-channel-outbound-policy-card"}
            >
              <div className={styles.heading}>
                {t.channelPage.policyDetails.outbound} {t.channelPage.policyDetails.heading}
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.feeRate}</div>
                <div className={classNames(styles.rowValue)}>{channelDetail?.feeRateMilliMsat}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.baseFee}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.feeBase : 0)}
                </div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.timeLockDelta}</div>
                <div className={classNames(styles.rowValue)}>{channelDetail?.remoteTimeLockDelta}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.minimumHTLC}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.minHtlc : 0)}
                </div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.maximumHTLC}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.maxHtlc : 0)}
                </div>
              </div>
            </div>
            <div
              className={classNames(styles.card, styles.currentCard)}
              data-intercom-target={"inspect-channel-inbound-policy-card"}
            >
              <div className={styles.heading}>
                {t.channelPage.policyDetails.inbound} {t.channelPage.policyDetails.heading}
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.feeRate}</div>
                <div className={classNames(styles.rowValue)}>{channelDetail?.remoteFeeRateMilliMsat}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.baseFee}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.remoteFeeBase : 0)}
                </div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.timeLockDelta}</div>
                <div className={classNames(styles.rowValue)}>{channelDetail?.remoteTimeLockDelta}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.minimumHTLC}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.remoteMinHtlc : 0)}
                </div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>{t.channelPage.policyDetails.maximumHTLC}</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.0f")(channelDetail ? channelDetail.remoteMaxHtlc : 0)}
                </div>
              </div>
            </div>
          </div>

          <div
            className={classNames(styles.card, styles.channelSummaryChart)}
            data-intercom-target={"inspect-channel-summary-chart-card"}
          >
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  intercomTarget={"inspect-channel-summary-chart-key-select"}
                  sizeVariant={InputSizeVariant.small}
                  value={profitKey}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateProfitChartKey({
                          key: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: t.channelPage.chart.amount,
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: t.channelPage.chart.amount },
                    { value: "revenue", label: t.channelPage.chart.revenue },
                    { value: "count", label: t.channelPage.chart.count },
                  ]}
                />
              </div>
              {/*<div className={styles.profitChartRightControls}>*/}
              {/*  <SettingsIcon />*/}
              {/*  Settings*/}
              {/*</div>*/}
            </div>
            <div className={styles.chartContainer}>
              {history && <ProfitsChart data={history.history} from={from} to={to} />}
            </div>
          </div>
        </div>

        <div className={classNames(styles.pageRow, styles.tripleRow)}>
          <div className={styles.card} data-intercom-target={"inspect-channel-capacity-card"}>
            <div className={styles.heading}>{t.channelPage.amount.header}</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.capacity}</div>
              <div className={styles.rowValue}>{ft(totalCapacity)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.outbound}</div>
              <div className={styles.rowValue}>{ft(historyAmountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.inbound}</div>
              <div className={classNames(styles.rowValue)}>{ft(historyAmountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.total}</div>
              <div className={styles.rowValue}>{ft(historyAmountTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.turnover}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.2")(historyAmountTotalCapacity)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.amount.balanceScore}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(".1%")(historyAmountInOut)}</div>
            </div>
          </div>

          <div className={styles.card} data-intercom-target={"inspect-channel-revenue-details-card"}>
            <div className={styles.heading}>{t.channelPage.revenue.header}</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.outbound}</div>
              <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.inbound}</div>
              <div className={classNames(styles.rowValue)}>{ft(historyRevenueIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.total}</div>
              <div className={styles.rowValue}>{ft(historyRevenueTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.averageFeeOut}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueOutAmountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.averageFeeIn}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueInAmountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.revenue.averageFee}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueTotalAmountTotal)}</div>
            </div>
          </div>

          <div className={styles.card} data-intercom-target={"inspect-channel-transaction-count-details-card"}>
            <div className={styles.heading}>{t.channelPage.transactionCount.header}</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.transactionCount.outbound}</div>
              <div className={styles.rowValue}>{ft(historyCountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.transactionCount.inbound}</div>
              <div className={classNames(styles.rowValue)}>{ft(historyCountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.transactionCount.total}</div>
              <div className={styles.rowValue}>{ft(historyCountTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>{t.channelPage.transactionCount.balanceScore}</div>
              <div className={classNames(styles.rowValue)}>{d3.format(".1%")(historyCountInOut)}</div>
            </div>
          </div>
        </div>

        <div
          className={classNames(styles.pageRow, styles.eventSummary)}
          data-intercom-target={"inspect-channel-event-summary-card"}
        >
          <div className={styles.shortColumn} data-intercom-target={"inspect-channel-event-summary-card"}>
            <EventsCard
              events={event as ChannelEventResponse}
              selectedEvents={selectedEvents}
              channels={history as ChannelHistoryResponse}
            />
          </div>
          <div
            className={classNames(styles.card, styles.channelSummaryChart)}
            style={{ height: "600px" }}
            data-intercom-target={"inspect-channel-event-summary-chart-card"}
          >
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  intercomTarget={"inspect-channel-event-summary-chart-key"}
                  value={eventKey}
                  sizeVariant={InputSizeVariant.small}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateEventChartKey({
                          key: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: t.channelPage.chart.amount,
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: t.channelPage.chart.amount },
                    { value: "revenue", label: t.channelPage.chart.revenue },
                    { value: "count", label: t.channelPage.chart.count },
                  ]}
                />
              </div>
              <div className={styles.profitChartRightControls}>
                <Popover
                  intercomTarget={"inspect-channel-event-toggle"}
                  button={
                    <Button
                      intercomTarget={"inspect-channel-event-toggle-button"}
                      buttonColor={ColorVariant.primary}
                      buttonSize={SizeVariant.small}
                      icon={<EventFlagIcon />}
                      hideMobileText={true}
                    >
                      {selectedEventsCount}
                    </Button>
                  }
                  className={"right"}
                >
                  <div className={styles.channelChartSettingsPopover}>
                    <div className={styles.cardRow}>
                      <div className={styles.rowLabel}>
                        <Switch
                          intercomTarget={"inspect-channel-event-toggle-all"}
                          label={t.channelPage.chartSetting.toggleAll}
                          checked={allToggle}
                          colorVariant={InputColorVaraint.accent1}
                          onChange={() => {
                            setAllToggle(!allToggle);
                            setSelectedEvents({
                              feeRate: !allToggle,
                              baseFee: !allToggle,
                              minHtlc: !allToggle,
                              maxHtlc: !allToggle,
                              enabled: !allToggle,
                              disabled: !allToggle,
                            });
                          }}
                        />
                      </div>
                    </div>

                    {Object.keys(selectedEvents).map((key) => {
                      const k = key as keyof typeof selectedEvents;
                      return (
                        <div className={styles.cardRow} key={key}>
                          <div className={styles.rowLabel}>
                            <Switch
                              intercomTarget={`inspect-channel-event-toggle-${key}`}
                              label={eventNames.get(k) || ""}
                              checked={selectedEvents[k] || false}
                              onChange={() => {
                                setSelectedEvents({ ...selectedEvents, [key]: !selectedEvents[k] });
                              }}
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </Popover>
              </div>
            </div>

            <div className={styles.chartContainer} data-intercom-target={"inspect-channel-event-chart"}>
              {history && (
                <EventsChart
                  from={from}
                  to={to}
                  data={history.history}
                  events={event as ChannelEventResponse}
                  selectedEventTypes={selectedEvents}
                />
              )}
            </div>
          </div>
        </div>

        <div className={classNames(styles.pageRow, styles.balanceChartWrapper)}>
          <div className={styles.card}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  intercomTarget={"inspect-channel-balance-chart-key"}
                  value={balanceChanId}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateBalanceChanID({
                          key: (newValue as { value: string; label: string }) || {
                            value: 0,
                            label: balance?.channelBalances ? balance.channelBalances[0]?.channelId.toString() : "",
                          },
                        })
                      );
                    }
                  }}
                  isDisabled={channelBalanceOptions.length < 2}
                  options={channelBalanceOptions}
                />
              </div>
            </div>
            <div className={classNames(styles.chartContainer)}>
              {!isLoading && balance?.channelBalances?.length && (
                <BalanceChart
                  data={balance.channelBalances[balanceChanId.value].balances}
                  totalCapacity={historyCapacity}
                  from={from}
                  to={to}
                />
              )}
            </div>
          </div>
        </div>

        <div className={styles.pageRow}>
          <div className={styles.card} data-intercom-target={"inspect-channel-flow-chart-card"}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  intercomTarget={"inspect-channel-flow-chart-key"}
                  value={flowKey}
                  sizeVariant={InputSizeVariant.small}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateFlowKey({
                          flowKey: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: t.channelPage.chart.amount,
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: t.channelPage.chart.amount },
                    { value: "revenue", label: t.channelPage.chart.revenue },
                    { value: "count", label: t.channelPage.chart.count },
                  ]}
                />
              </div>
              <div className={styles.profitChartRightControls}>
                {/*<Popover*/}
                {/*  button={<Button text={"Settings"} icon={<SettingsIcon />} hideMobileText={true} />}*/}
                {/*  className={"right"}*/}
                {/*>*/}
                {/*  Hello*/}
                {/*</Popover>*/}
              </div>
            </div>
            <div className="legendsContainer">
              <div className="sources">{t.channelPage.legendsContainer.sources}</div>
              <div className="outbound">{t.channelPage.legendsContainer.outbound}</div>
              <div className="inbound">{t.channelPage.legendsContainer.inbound}</div>
              <div className="destinations">{t.channelPage.legendsContainer.destinations}</div>
            </div>
            <div className={classNames(styles.chartWrapper, styles.flowChartWrapper)}>
              {!isLoading && data && <FlowChart data={data} />}
            </div>
          </div>
        </div>
      </div>
    </PopoutPageTemplate>
  );
}

export default ChannelPage;
