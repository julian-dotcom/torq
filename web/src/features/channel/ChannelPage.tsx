import styles from "./channel-page.module.scss";
import classNames from "classnames";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import BarChart from "./charts/barChart/BarChart";
import Switch from "../inputs/Slider/Switch";

const data = [
  { date: new Date(2022, 1, 2), revenue: 250.6, sales: 10979000 },
  { date: new Date(2022, 1, 3), revenue: 231, sales: 9303000 },
  { date: new Date(2022, 1, 4), revenue: 280.4, sales: 8185000 },
  { date: new Date(2022, 1, 5), revenue: 270.9, sales: 8213000 },
  { date: new Date(2022, 1, 6), revenue: 280.4, sales: 8518000 },
  { date: new Date(2022, 1, 7), revenue: 250.3, sales: 8991000 },
  { date: new Date(2022, 1, 8), revenue: 280.6, sales: 8620000 },
  { date: new Date(2022, 1, 9), revenue: 210.5, sales: 8479000 },
  { date: new Date(2022, 1, 10), revenue: 280.5, sales: 8479000 },
  { date: new Date(2022, 1, 11), revenue: 210.7, sales: 8217000 },
  { date: new Date(2022, 1, 12), revenue: 280.8, sales: 8085000 },
  { date: new Date(2022, 1, 13), revenue: 220.3, sales: 8638000 },
  { date: new Date(2022, 1, 14), revenue: 280.5, sales: 8778000 },
  { date: new Date(2022, 1, 15), revenue: 280.8, sales: 8352000 },
];

function ChannelPage() {
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
            {/*<div className={styles.card}>Automation</div>*/}
          </div>

          <div className={classNames(styles.card, styles.channelSummaryChart)}>
            <BarChart data={data} />
          </div>
        </div>
        <div className={styles.pageRow}>
          <div className={styles.card}>
            Channel Events
            <div className={styles.chartWrapper} style={{ height: "500px" }}>
              <BarChart data={data} />
            </div>
          </div>
        </div>
        <div className={styles.pageRow}>
          <div className={styles.card}>
            Channel Flow
            <div className={styles.chartWrapper} style={{ height: "500px" }}>
              <BarChart data={data} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ChannelPage;
