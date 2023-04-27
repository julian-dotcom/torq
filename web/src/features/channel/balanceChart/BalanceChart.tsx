// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "features/charts/useD3";
import * as d3 from "d3";
import { useEffect } from "react";
import { NumberValue, Selection } from "d3";
import ChartCanvas from "features/charts/chartCanvas";
import "features/charts/chart.scss";
import { AreaPlot } from "features/charts/charts";
import clone from "clone";
import { useGetSettingsQuery } from "apiSlice";

type BalanceChart = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any[];
  totalCapacity: number;
  from: string;
  to: string;
};

function BalanceChart({ data, totalCapacity, from, to }: BalanceChart) {
  let chart: ChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];
  const settings = useGetSettingsQuery();

  // Check and update the chart size if the navigation changes the container size
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const navCheck = (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
    return () => {
      const boundingBox = container?.node()?.getBoundingClientRect();
      if (currentSize[0] !== boundingBox?.width || currentSize[1] !== boundingBox?.height) {
        chart.resizeChart();
        chart.draw();
        currentSize = [boundingBox?.width, boundingBox?.height];
      }
    };
  };

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let paddedData: any[] = [];

      if (data?.length > 0) {
        paddedData = clone(data);
        paddedData.unshift({
          capacity_diff: 0,
          date: from,
          inboundCapacity: data[0].inboundCapacity,
          outboundCapacity: data[0].outboundCapacity,
        });
        paddedData.push({
          capacity_diff: 0,
          date: to,
          inboundCapacity: data[Math.max(data.length - 1, 0)].inboundCapacity,
          outboundCapacity: data[Math.max(data.length - 1, 0)].outboundCapacity,
        });
      }

      chart = new ChartCanvas(container, paddedData, {
        from: new Date(from),
        to: new Date(to),
        timezone: settings?.data?.preferredTimezone || "UTC",
        yScaleKey: "outboundCapacity",
        rightYScaleKey: "outboundCapacity",
        rightYAxisKeys: ["outboundCapacity"],
        yAxisMaxOverride: totalCapacity,
        rightYAxisMaxOverride: totalCapacity,
        xAxisLabelFormatter: d3.timeFormat("%d %b - %H:%M") as (domainValue: NumberValue) => string,
      });

      chart.plot(AreaPlot, {
        id: "outboundCapacity",
        key: "outboundCapacity",
        legendLabel: "Outbound capacity",
        curveFunction: d3.curveStepAfter,
        areaGradient: ["rgba(133, 196, 255, 0.5)", "rgba(87, 211, 205, 0.6)"],
        // areaColor: "#FAAE93",
      });
      chart.draw();

      setInterval(navCheck(container), 200);
    },
    [data, data ? data[0].date : "", data ? data[data.length - 1].date : ""]
  );

  useEffect(() => {
    return () => {
      if (chart) {
        chart.removeResizeListener();
      }
    };
  }, [data]);

  return <div ref={ref} className={"chart-ref"} />;
}

export default BalanceChart;
