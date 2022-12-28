import { Flag16Regular as EventFlagIcon } from "@fluentui/react-icons";
import {
  useGetChannelHistoryQuery,
  useGetChannelBalanceQuery,
  useGetChannelOnChainCostQuery,
  useGetChannelRebalancingQuery,
  useGetChannelEventQuery,
  useGetFlowQuery,
} from "apiSlice";
import { ChannelHistoryResponse, ChannelEventResponse, Channel } from "features/channel/channelTypes";
import type { GetChannelHistoryData, GetFlowQueryParams } from "types/api";
import classNames from "classnames";
import * as d3 from "d3";
import { addDays, format } from "date-fns";
import DetailsPageTemplate from "features/templates/detailsPageTemplate/DetailsPageTemplate";
import React from "react";
import { useParams } from "react-router";
import { Link } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "store/hooks";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import EventsCard from "features/eventsCard/EventsCard";
import Select from "components/forms/select/Select";
import Switch from "components/forms/switch/Switch";
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

const ft = d3.format(",.0f");

const eventNames = new Map([
  ["fee_rate", "Fee rate"],
  ["base_fee", "Base fee"],
  ["min_htlc", "Min htlc"],
  ["max_htlc", "Max htlc"],
  ["enabled", "Enabled"],
  ["disabled", "Disabled"],
]);

type ChannelPageProps = {
  chanId?: string;
};

function ChannelPage(_: ChannelPageProps) {
  const currentPeriod = useAppSelector(selectTimeInterval);
  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [allToggle, setAllToggle] = React.useState(true);
  const [selectedEvents, setSelectedEvents] = React.useState(
    new Map<string, boolean>([
      ["fee_rate", true],
      ["base_fee", true],
      ["min_htlc", true],
      ["max_htlc", true],
      ["enabled", true],
      ["disabled", true],
    ])
  );
  const handleSelectEventUpdate = (type: string) => {
    return (checked: boolean) => {
      setSelectedEvents(new Map(selectedEvents.set(type, checked)));
    };
  };

  const { chanId } = useParams();
  const flowQueryParams: GetFlowQueryParams = {
    from: from,
    to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
    chanIds: chanId || "1",
    network: activeNetwork,
  };
  const { data, isLoading } = useGetFlowQuery(flowQueryParams);

  const channelHistoryQueryData: GetChannelHistoryData = {
    params: { chanId: chanId || "1" },
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
    balanceChanId = { value: 0, label: balance?.channelBalances ? balance.channelBalances[0]?.lndShortChannelId : "" };
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

  const selectedEventsCount = Array.from(selectedEvents).filter((d) => d[1]).length;
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
  const historyCapacity = history?.channels.length ? history?.channels[balanceChanId.value].capacity : 0;

  const title =
    !isLoading &&
    history &&
    (history.channels || [])
      .map((d: Channel, _: number) => {
        return d.alias;
      })
      .filter((value: any, index: number, self: any[]) => {
        return self.indexOf(value) === index;
      })
      .join(", ");
  let channelBalanceOptions = [{ value: 0, label: "" }];
  if (balance?.channelBalances) {
    channelBalanceOptions = balance.channelBalances.map((d: any, i: number) => {
      return { value: i, label: d.ChanId };
    });
  }
  const breadcrumbs = [
    <span key="b1">Analyse;</span>,
    <span key="b2">Inspect</span>,
    <Link key="b3" to={`/analyse/inspect/${chanId}`}>
      {chanId}
    </Link>,
  ];
  return (
    <DetailsPageTemplate title={title as string} titleContent={<TimeIntervalSelect />} breadcrumbs={breadcrumbs}>
      <div className={styles.channelWrapper}>
        <div className={classNames(styles.pageRow, styles.channelSummary)}>
          <div className={styles.shortColumn}>
            <div className={styles.card}>
              <div className={styles.heading}>Revenue</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Forwarding fees</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Channel Leases</div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>{history?.revenueOut}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Expenses</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Rebalancing</div>
                <div className={classNames(styles.rowValue)}>{ft(rebalancingCostBy1000)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Open & Close</div>
                <div className={classNames(styles.rowValue)}>{ft(onchainCost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={classNames(styles.rowValue)}>{ft(onchainPlusRebalancingCostBy1000)}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Profit</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={classNames(styles.rowValue)}>{ft(profit)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Gross Profit Margin</div>
                <div className={classNames(styles.rowValue)}>{d3.format(".2%")(historyReveueOutMinusCost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Turnover</div>
                <div className={classNames(styles.rowValue)}>{d3.format(",.2")(historyAmountTotalCapacity)}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Automation</div>
              <div className={styles.sliderRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"Fees"} />
                </div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.sliderRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"Rebalancing"} />
                </div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.sliderRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"HTLC amount"} />
                </div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
            </div>
          </div>

          <div className={classNames(styles.card, styles.channelSummaryChart)}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  className={"small"}
                  value={profitKey}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateProfitChartKey({
                          key: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: "Amount",
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: "Amount" },
                    { value: "revenue", label: "Revenue" },
                    { value: "count", label: "Count" },
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
          <div className={styles.card}>
            <div className={styles.heading}>Amount</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Capacity</div>
              <div className={styles.rowValue}>{ft(totalCapacity)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Outbound</div>
              <div className={styles.rowValue}>{ft(historyAmountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Inbound</div>
              <div className={classNames(styles.rowValue)}>{ft(historyAmountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Total</div>
              <div className={styles.rowValue}>{ft(historyAmountTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Turnover</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.2")(historyAmountTotalCapacity)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Balance score</div>
              <div className={classNames(styles.rowValue)}>{d3.format(".1%")(historyAmountInOut)}</div>
            </div>
          </div>

          <div className={styles.card}>
            <div className={styles.heading}>Revenue</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Outbound</div>
              <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Inbound</div>
              <div className={classNames(styles.rowValue)}>{ft(historyRevenueIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Total</div>
              <div className={styles.rowValue}>{ft(historyRevenueTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Average fee out</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueOutAmountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Average fee in</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueInAmountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Average fee</div>
              <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueTotalAmountTotal)}</div>
            </div>
          </div>

          <div className={styles.card}>
            <div className={styles.heading}>Transaction count</div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Outbound</div>
              <div className={styles.rowValue}>{ft(historyCountOut)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Inbound</div>
              <div className={classNames(styles.rowValue)}>{ft(historyCountIn)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Total</div>
              <div className={styles.rowValue}>{ft(historyCountTotal)}</div>
            </div>
            <div className={styles.cardRow}>
              <div className={styles.rowLabel}>Balance score (Nb. Tx)</div>
              <div className={classNames(styles.rowValue)}>{d3.format(".1%")(historyCountInOut)}</div>
            </div>
          </div>
        </div>

        <div className={classNames(styles.pageRow, styles.eventSummary)}>
          <div className={styles.shortColumn}>
            <EventsCard
              events={event as ChannelEventResponse}
              selectedEvents={selectedEvents}
              channels={history as ChannelHistoryResponse}
            />
          </div>
          <div className={classNames(styles.card, styles.channelSummaryChart)} style={{ height: "600px" }}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  value={eventKey}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateEventChartKey({
                          key: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: "Amount",
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: "Amount" },
                    { value: "revenue", label: "Revenue" },
                    { value: "count", label: "Count" },
                  ]}
                />
              </div>
              <div className={styles.profitChartRightControls}>
                <Popover
                  button={
                    <Button
                      buttonColor={ColorVariant.primary}
                      buttonSize={SizeVariant.small}
                      isOpen={selectedEventsCount > 0}
                      icon={<EventFlagIcon />}
                      className={"collapse-tablet"}
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
                          label="Toggle all"
                          checked={allToggle}
                          onChange={() => {
                            setAllToggle(!allToggle);
                            setSelectedEvents(
                              new Map([
                                ["fee_rate", !allToggle],
                                ["base_fee", !allToggle],
                                ["min_htlc", !allToggle],
                                ["max_htlc", !allToggle],
                                ["enabled", !allToggle],
                                ["disabled", !allToggle],
                              ])
                            );
                          }}
                        />
                      </div>
                    </div>

                    {Array.from(selectedEvents).map((item) => {
                      return (
                        <div className={styles.cardRow} key={item[0]}>
                          <div className={styles.rowLabel}>
                            <Switch
                              label={eventNames.get(item[0]) || ""}
                              checked={selectedEvents.get(item[0])}
                              onChange={(e) => handleSelectEventUpdate(item[0])}
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </Popover>
              </div>
            </div>

            <div className={styles.chartContainer}>
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
                  value={balanceChanId}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateBalanceChanID({
                          key: (newValue as { value: string; label: string }) || {
                            value: 0,
                            label: balance?.channelBalances ? balance.channelBalances[0]?.lndShortChannelId : "",
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
          <div className={styles.card}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  value={flowKey}
                  onChange={(newValue) => {
                    if (newValue) {
                      dispatch(
                        updateFlowKey({
                          flowKey: (newValue as { value: string; label: string }) || {
                            value: "amount",
                            label: "Amount",
                          },
                        })
                      );
                    }
                  }}
                  options={[
                    { value: "amount", label: "Amount" },
                    { value: "revenue", label: "Revenue" },
                    { value: "count", label: "Count" },
                  ]}
                />
              </div>
              <div className={styles.profitChartRightControls}>
                {/*<Popover*/}
                {/*  button={<Button text={"Settings"} icon={<SettingsIcon />} className={"collapse-tablet"} />}*/}
                {/*  className={"right"}*/}
                {/*>*/}
                {/*  Hello*/}
                {/*</Popover>*/}
              </div>
            </div>
            <div className="legendsContainer">
              <div className="sources">Sources</div>
              <div className="outbound">Outbound</div>
              <div className="inbound">Inbound</div>
              <div className="destinations">Destinations</div>
            </div>
            <div className={classNames(styles.chartWrapper, styles.flowChartWrapper)}>
              {!isLoading && data && <FlowChart data={data} />}
            </div>
          </div>
        </div>
      </div>
    </DetailsPageTemplate>
  );
}

export default ChannelPage;
