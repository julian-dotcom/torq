import styles from "./channel-page.module.scss";
import classNames from "classnames";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import ProfitsChart from "./revenueChart/ProtifsChart";
import EventsChart from "./eventsChart/EventsChart";
import Switch from "../inputs/Slider/Switch";
import Button from "../buttons/Button";
import Select from "../inputs/Select";
import { Settings16Regular as SettingsIcon } from "@fluentui/react-icons";
import FlowChart from "./flowChart/FlowChart";

import { useGetFlowQuery, useGetChannelHistoryQuery } from "apiSlice";
import { useAppSelector, useAppDispatch } from "../../store/hooks";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useParams } from "react-router";
import { selectFlowKeys, updateFlowKey } from "./channelSlice";

const chanData = [
  {
    date: new Date(2022, 11, 1),
    revenue: 2910,
    capacity_out: 2503000,
    events: [
      { type: "channel_open", id: 3 },
      { type: "fee_rate", value: 450, id: 4 },
      { type: "base_fee", value: 1, id: 4 },
    ],
  },
  { date: new Date(2022, 11, 2), revenue: 2310, capacity_out: 2803000 },
  {
    date: new Date(2022, 11, 3),
    revenue: 2510,
    capacity_out: 3003000,
    events: [{ type: "fee_rate", value: 245, id: 2 }],
  },
  { date: new Date(2022, 11, 4), revenue: 3800, capacity_out: 2585000 },
  {
    date: new Date(2022, 11, 5),
    revenue: 3000,
    capacity_out: 2213000,
  },
  { date: new Date(2022, 11, 6), revenue: 4812, capacity_out: 2518000 },
  {
    date: new Date(2022, 11, 7),
    revenue: 3500,
    capacity_out: 2591000,
    events: [
      { type: "rebalanced_out", value: 1234500, id: 5 },
      { type: "fee_rate", value: 355, id: 5 },
      { type: "fee_rate", value: 230, id: 5 },
      { type: "fee_rate", value: 412, id: 5 },
      { type: "fee_rate", value: 150, id: 5 },
    ],
  },
  { date: new Date(2022, 11, 8), revenue: 3800, capacity_out: 2020000 },
  {
    date: new Date(2022, 11, 9),
    revenue: 4100,
    capacity_out: 1879000,
  },
  { date: new Date(2022, 11, 10), revenue: 3780, capacity_out: 1579000 },
  { date: new Date(2022, 11, 11), revenue: 3150, capacity_out: 2317000 },
  {
    date: new Date(2022, 11, 12),
    revenue: 3980,
    capacity_out: 2785000,
  },
  {
    date: new Date(2022, 11, 13),
    revenue: 4220,
    capacity_out: 3538000,
    events: [{ type: "channel_status_disabled", id: 8 }],
  },
  { date: new Date(2022, 11, 14), revenue: 4980, capacity_out: 4578000 },
  {
    date: new Date(2022, 11, 15),
    revenue: 5280,
    capacity_out: 5352000,
    events: [{ type: "channel_status_enabled", id: 10 }],
  },
  {
    date: new Date(2022, 11, 16),
    revenue: 5280,
    capacity_out: 4352000,
    events: [
      { type: "fee_rate", value: 550, id: 10 },
      { type: "rebalanced_in", value: 423145, id: 10 },
    ],
  },
  { date: new Date(2022, 11, 17), revenue: 5280, capacity_out: 3252000 },
  {
    date: new Date(2022, 11, 18),
    revenue: 3280,
    capacity_out: 3852000,
    events: [{ type: "channel_close", id: 10 }],
  },
];

function ChannelPage() {
  const currentPeriod = useAppSelector(selectTimeInterval);
  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");
  let { chanId } = useParams();
  const { data, isLoading } = useGetFlowQuery({
    from: from,
    to: to,
    chanId: chanId || " ",
  });
  const historyQuery = useGetChannelHistoryQuery({
    chanIds: chanId || " ",
  });

  const flowKey = useAppSelector(selectFlowKeys);

  return (
    <div className={styles.channelsPageContent}>
      <div className={styles.channelControls}>
        <div className={styles.leftContainer}>
          <div className={styles.upperContainer}>Seetee Node</div>
          <div className={styles.lowerContainer}></div>
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
                <div className={styles.rowValue}>5,830</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Channel Leases</div>
                <div className={styles.rowValue}>501</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>6,331</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Expenses</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Rebalansing</div>
                <div className={styles.rowValue}>503</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Open & Close</div>
                <div className={styles.rowValue}>450</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>953</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Profit</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>5,378</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>APY</div>
                <div className={styles.rowValue}>11.2%</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Turnover</div>
                <div className={styles.rowValue}>3.14</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Automation</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"Fees"} />
                </div>
                <div className={styles.rowValue}>Progressive</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"Rebalancing"} />
                </div>
                <div className={styles.rowValue}>Custom</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>
                  <Switch label={"HTLC amount"} />
                </div>
                <div className={styles.rowValue}>Conservative</div>
              </div>
            </div>
          </div>

          <div className={classNames(styles.card, styles.channelSummaryChart)}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Button text={"Revenue"} isOpen={true} />
              </div>
              <div className={styles.profitChartRightControls}>
                <SettingsIcon />
                Settings
              </div>
            </div>
            <div className={styles.chartContainer}>
              {historyQuery.data && <ProfitsChart data={historyQuery.data.data} />}
            </div>
          </div>
        </div>

        <div className={styles.pageRow}>
          <div className={classNames(styles.card, styles.channelSummaryChart)} style={{ height: "600px" }}>
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Button text={"Revenue"} isOpen={true} />
              </div>
              <div className={styles.profitChartRightControls}>
                <SettingsIcon />
                Settings
              </div>
            </div>

            <div className={styles.chartContainer}>
              {/*<EventsChart data={chanData} />*/}
              {historyQuery.data && <EventsChart data={historyQuery.data.data} />}
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
              {/*<div className={styles.profitChartRightControls}>*/}
              {/*  <SettingsIcon />*/}
              {/*  Settings*/}
              {/*</div>*/}
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
