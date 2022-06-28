import React from "react";
import styles from "./channel-page.module.scss";
import * as d3 from "d3";
import classNames from "classnames";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import ProfitsChart from "./revenueChart/ProfitsChart";
import Button from "../buttons/Button";
import Select from "../inputs/Select";
import { Navigation20Regular as NavigationIcon, Gauge20Regular as DashboardIcon } from "@fluentui/react-icons";
import FlowChart from "./flowChart/FlowChart";
import { toggleNav } from "../navigation/navSlice";
import { useGetFlowQuery, useGetChannelHistoryQuery } from "apiSlice";
import { useAppSelector, useAppDispatch } from "../../store/hooks";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useParams } from "react-router";
import {
  selectEventChartKey,
  selectFlowKeys,
  selectProfitChartKey,
  updateFlowKey,
  updateProfitChartKey,
} from "./channelSlice";

const ft = d3.format(",.0f");

const eventNames = new Map([
  ["fee_rate", "Fee rate"],
  ["base_fee", "Base fee"],
  ["min_htlc", "Min htlc"],
  ["max_htlc", "Max htlc"],
  ["enabled", "Enabled"],
  ["disabled", "Disabled"],
]);

function ChannelPage() {
  const currentPeriod = useAppSelector(selectTimeInterval);
  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");
  let [selectedEvents, setSelectedEvents] = React.useState(
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
    return (e: React.ChangeEvent<HTMLInputElement>) => {
      setSelectedEvents(new Map(selectedEvents.set(type, e.target.checked)));
    };
  };

  let { chanId } = useParams();
  const { data, isLoading } = useGetFlowQuery({
    from: from,
    to: to,
    chanId: chanId || "1",
  });
  const historyQuery = useGetChannelHistoryQuery({
    from: from,
    to: to,
    chanIds: chanId || "1",
  });

  const flowKey = useAppSelector(selectFlowKeys);
  const profitKey = useAppSelector(selectProfitChartKey);
  const eventKey = useAppSelector(selectEventChartKey);

  let total_capacity: number = 0;
  if (historyQuery?.data?.channels) {
    total_capacity = historyQuery.data.channels
      .map((d: { capacity: number }) => {
        return d.capacity;
      })
      .reduce((partialSum: number, a: number) => partialSum + a, 0);
  }

  const selectedEventsCount = Array.from(selectedEvents).filter((d) => d[1]).length;
  const profit: number =
    historyQuery?.data?.revenue_out - historyQuery?.data?.on_chain_cost - historyQuery?.data?.rebalancing_cost / 1000;

  const totalCost: number = historyQuery?.data?.on_chain_cost + historyQuery?.data?.rebalancing_cost / 1000;

  return (
    <div className={styles.channelsPageContent}>
      <div className={styles.channelControls}>
        <div className={styles.leftContainer}>
          <div className={styles.lowerContainer}>
            <Button
              icon={<NavigationIcon />}
              text={"Menu"}
              onClick={() => dispatch(toggleNav())}
              className={"show-nav-btn collapse-tablet"}
            />
            <Button icon={<DashboardIcon />} text={"Dashboard"} />
          </div>
        </div>
        <div className={styles.rightContainer}>
          <TimeIntervalSelect />
        </div>
      </div>

      <div className={styles.channelWrapper}>
        <div className={classNames(styles.pageRow, styles.channelSummary)}>
          <div className={styles.shortColumn}>
            <div className={styles.card}>
              <div className={styles.heading}>Revenue</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Forwarding fees</div>
                <div className={styles.rowValue}>{ft(historyQuery?.data?.revenue_out)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Channel Leases</div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>{ft(historyQuery?.data?.revenue_out)}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Expenses</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Rebalancing</div>
                <div className={classNames(styles.rowValue)}>{ft(historyQuery?.data?.rebalancing_cost / 1000)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Open & Close</div>
                <div className={classNames(styles.rowValue)}>{ft(historyQuery?.data?.on_chain_cost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={classNames(styles.rowValue)}>{ft(totalCost)}</div>
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
                <div className={classNames(styles.rowValue)}>
                  {d3.format(".2%")((historyQuery?.data?.revenue_out - totalCost) / historyQuery?.data?.revenue_out)}
                </div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Revenue</div>
                <div className={styles.rowValue}>{ft(historyQuery?.data?.revenue_out)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Amount</div>
                <div className={classNames(styles.rowValue)}>{ft(historyQuery?.data?.amount_out)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Transactions</div>
                <div className={styles.rowValue}>{ft(historyQuery?.data?.count_out)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Average fee</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.1f")((historyQuery?.data?.revenue_out / historyQuery?.data?.amount_out) * 1000 * 1000)}
                </div>
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
              {historyQuery.data && <ProfitsChart data={historyQuery.data.history} dashboard={true} />}
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
    </div>
  );
}

export default ChannelPage;
