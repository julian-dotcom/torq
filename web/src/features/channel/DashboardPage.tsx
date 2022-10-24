import {
  useGetChannelHistoryQuery,
  useGetChannelRebalancingQuery,
  useGetChannelOnChainCostQuery,
  useGetFlowQuery
} from "apiSlice";
import {
  ChannelOnchainCostResponse,
  ChannelHistoryResponse,
  ChannelRebalancingResponse,
} from "features/channel/channelTypes"
import type { GetChannelHistoryQueryParams } from "types/api";
import classNames from "classnames";
import * as d3 from "d3";
import { addDays, format } from "date-fns";
import DetailsPageTemplate from "features/templates/detailsPageTemplate/DetailsPageTemplate";
import { useParams } from "react-router";
import { Link, Outlet } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "store/hooks";
import Select from "features/inputs/Select";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import styles from "./channel-page.module.scss";
import { selectFlowKeys, selectProfitChartKey, updateFlowKey, updateProfitChartKey } from "./channelSlice";
import FlowChart from "./flowChart/FlowChart";
import ProfitsChart from "./revenueChart/ProfitsChart";
import clone from "clone";

const ft = d3.format(",.0f");

function ChannelPage() {
  const currentPeriod = useAppSelector(selectTimeInterval);

  const dispatch = useAppDispatch();
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  const { chanId } = useParams();

  const queryParams: GetChannelHistoryQueryParams = {
    from: from,
    to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
    chanIds: chanId || "1",
  }

  const { data, isLoading } = useGetFlowQuery(queryParams);
  const { data: onChainCostResponse } = useGetChannelOnChainCostQuery(queryParams);
  const onChainCost = clone<ChannelOnchainCostResponse>(onChainCostResponse as ChannelOnchainCostResponse) || 0;

  const { data: historyQueryResponse } = useGetChannelHistoryQuery(queryParams);
  const historyQuery = clone<ChannelHistoryResponse>(historyQueryResponse as ChannelHistoryResponse) || null;

  const { data: rebalancingResponse } = useGetChannelRebalancingQuery(queryParams);
  const rebalancing = clone<ChannelRebalancingResponse>(rebalancingResponse as ChannelRebalancingResponse) || null;

  const flowKey = useAppSelector(selectFlowKeys);
  const profitKey = useAppSelector(selectProfitChartKey);

  const profit: number =
    historyQuery?.revenueOut - onChainCost?.onChainCost - rebalancing?.rebalancingCost / 1000;

  const totalCost: number = onChainCost?.onChainCost + rebalancing?.rebalancingCost / 1000;

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"analyse"}>
      Summary
    </Link>,
  ];
  return (
    <DetailsPageTemplate title={"Forwards Summary"} titleContent={<TimeIntervalSelect />} breadcrumbs={breadcrumbs}>
      <div className={styles.channelWrapper}>
        <div className={classNames(styles.pageRow, styles.channelSummary)}>
          <div className={styles.shortColumn}>
            <div className={styles.card}>
              <div className={styles.heading}>Revenue</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Forwarding fees</div>
                <div className={styles.rowValue}>{ft(historyQuery?.revenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Channel Leases</div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>{ft(historyQuery?.revenueOut)}</div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.heading}>Expenses</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Rebalancing</div>
                <div className={classNames(styles.rowValue)}>{ft(rebalancing?.rebalancingCost / 1000)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Open & Close</div>
                <div className={classNames(styles.rowValue)}>{ft(onChainCost?.onChainCost)}</div>
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
                  {d3.format(".2%")((historyQuery?.revenueOut - totalCost) / historyQuery?.revenueOut)}
                </div>
              </div>
            </div>
            <div className={styles.card}>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Revenue</div>
                <div className={styles.rowValue}>{ft(historyQuery?.revenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Amount</div>
                <div className={classNames(styles.rowValue)}>{ft(historyQuery?.amountOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Transactions</div>
                <div className={styles.rowValue}>{ft(historyQuery?.countOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Average fee</div>
                <div className={classNames(styles.rowValue)}>
                  {d3.format(",.1f")((historyQuery?.revenueOut / historyQuery?.amountOut) * 1000 * 1000)}
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
              {historyQuery && (
                <ProfitsChart data={historyQuery.history} dashboard={true} from={from} to={to} />
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
      <Outlet />
    </DetailsPageTemplate>
  );
}

export default ChannelPage;
