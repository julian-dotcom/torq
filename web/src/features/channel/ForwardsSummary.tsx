import {
  useGetChannelHistoryQuery,
  useGetChannelRebalancingQuery,
  useGetChannelOnChainCostQuery,
  useGetFlowQuery,
} from "apiSlice";
import type { GetChannelHistoryData, GetFlowQueryParams } from "types/api";
import classNames from "classnames";
import * as d3 from "d3";
import { addDays, format } from "date-fns";
import DetailsPageTemplate from "features/templates/detailsPageTemplate/DetailsPageTemplate";
import { useParams } from "react-router";
import { Link, Outlet } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "store/hooks";
import Select from "components/forms/select/Select";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import styles from "./channel-page.module.scss";
import { selectFlowKeys, selectProfitChartKey, updateFlowKey, updateProfitChartKey } from "./channelSlice";
import FlowChart from "./flowChart/FlowChart";
import ProfitsChart from "./revenueChart/ProfitsChart";
import { selectActiveNetwork } from "features/network/networkSlice";
import { InputSizeVariant } from "components/forms/forms";

const ft = d3.format(",.0f");

function FowardsSummaryPage() {
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const currentPeriod = useAppSelector(selectTimeInterval);

  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  const { chanId } = useParams();

  const flowQueryParams: GetFlowQueryParams = {
    from: from,
    to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
    chanIds: chanId || "all",
    network: activeNetwork,
  };

  const { data, isLoading } = useGetFlowQuery(flowQueryParams);

  const getChannelHistoryData: GetChannelHistoryData = {
    params: { chanId: chanId || "all" },
    queryParams: {
      from: from,
      to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
      network: activeNetwork,
    },
  };

  const { data: onChainCost } = useGetChannelOnChainCostQuery(getChannelHistoryData);
  const { data: history } = useGetChannelHistoryQuery(getChannelHistoryData);
  const { data: rebalancing } = useGetChannelRebalancingQuery(getChannelHistoryData);

  const flowKey = useAppSelector(selectFlowKeys);
  const profitKey = useAppSelector(selectProfitChartKey);

  const profit: number =
    history?.revenueOut && onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? history?.revenueOut - onChainCost?.onChainCost - rebalancing?.rebalancingCost / 1000
      : 0;

  const totalCost: number =
    onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? onChainCost?.onChainCost + rebalancing?.rebalancingCost / 1000
      : 0;
  const historyRevenueOut = history?.revenueOut || 0;
  const rebalancingCostBy1000 = rebalancing?.rebalancingCost ? rebalancing?.rebalancingCost / 1000 : 0;
  const onchainCost = onChainCost?.onChainCost || 0;
  const historyReveueOutMinusCost = history?.revenueOut ? (history?.revenueOut - totalCost) / history?.revenueOut : 0;
  const historyAmountOut = history?.amountOut || 0;
  const historyCountOut = history?.countOut || 0;
  const historyRevenueOutAmountOut =
    history?.revenueOut && history?.amountOut ? (history?.revenueOut / history?.amountOut) * 1000 * 1000 : 0;

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"analyse"}>
      Summary
    </Link>,
  ];
  return (
    <DetailsPageTemplate title={"Forwards Summary"} titleContent={<TimeIntervalSelect />} breadcrumbs={breadcrumbs}>
      <div className={styles.channelWrapper}>
        <div
          className={classNames(styles.pageRow, styles.channelSummary)}
          data-intercom-target={"forwards-summary-container"}
        >
          <div className={styles.shortColumn} data-intercom-target={"forwards-summary-stats"}>
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
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
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
                <div className={classNames(styles.rowValue)}>{d3.format(".2%")(historyReveueOutMinusCost)}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Revenue</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Amount</div>
                <div className={classNames(styles.rowValue)}>{ft(historyAmountOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Transactions</div>
                <div className={styles.rowValue}>{ft(historyCountOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Average fee</div>
                <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueOutAmountOut)}</div>
              </div>
            </div>
          </div>

          <div
            className={classNames(styles.card, styles.channelSummaryChart)}
            data-intercom-target={"forwards-summary-chart"}
          >
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  sizeVariant={InputSizeVariant.small}
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
              {history && <ProfitsChart data={history.history} dashboard={true} from={from} to={to} />}
            </div>
          </div>
        </div>

        <div className={styles.pageRow} data-intercom-target={"forwards-summary-flow-chart"}>
          <div className={styles.card}>
            <div className={styles.profitChartControls}>
              <div
                className={styles.profitChartLeftControls}
                data-intercom-target={"forwards-summary-flow-chart-select"}
              >
                <Select
                  value={flowKey}
                  sizeVariant={InputSizeVariant.small}
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
                {/*  button={<Button text={"Settings"} icon={<SettingsIcon />} hideMobileText={true} />}*/}
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
      <Outlet />
    </DetailsPageTemplate>
  );
}

export default FowardsSummaryPage;
