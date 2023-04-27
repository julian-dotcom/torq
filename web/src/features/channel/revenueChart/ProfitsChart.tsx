// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "features/charts/useD3";
import { useEffect, useState } from "react";
import { Selection } from "d3";
import ChartCanvas from "features/charts/chartCanvas";
import "features/charts/chart.scss";
import { BarPlot, LinePlot } from "features/charts/charts";
import { useGetSettingsQuery } from "apiSlice";
import { AnyObject } from "utils/types";

type ProfitsChart = {
  data: AnyObject[];
  dataKey: string;
  dashboard?: boolean;
  from: string;
  to: string;
};

export const ProfitChartKeyOptions = [
  { value: "amount", label: "Amount" },
  { value: "revenue", label: "Revenue" },
  { value: "count", label: "Count" },
];

function ProfitsChart({ data, dashboard, dataKey, to, from }: ProfitsChart) {
  let chart: ChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];
  const settings = useGetSettingsQuery();
  const [label, setLabel] = useState<string>(ProfitChartKeyOptions[0].label);

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

  useEffect(() => {
    setLabel(ProfitChartKeyOptions.find((o) => o.value === dataKey)?.label || "Amount");
  }, [dataKey]);

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
      if (dashboard) {
        chart = new ChartCanvas(container, data, {
          from: new Date(from),
          to: new Date(to),
          timezone: settings?.data?.preferredTimezone || "UTC",
          yScaleKey: dataKey + "Out",
          rightYScaleKey: dataKey + "Out",
          rightYAxisKeys: [dataKey + "Out"],
          xAxisPadding: 12,
        });
        chart.plot(BarPlot, {
          id: dataKey + "Out",
          key: dataKey + "Out",
          legendLabel: label + " out",
          barColor: "rgba(186, 147, 250, 0.6)",
          barHoverColor: "rgba(186, 147, 250, 0.8)",
          // areaGradient: ["rgba(133, 196, 255, 0.5)", "rgba(87, 211, 205, 0.5)"],
        });
        chart.draw();
      } else {
        chart = new ChartCanvas(container, data, {
          from: new Date(from),
          to: new Date(to),
          timezone: settings?.data?.preferredTimezone || "UTC",
          yScaleKey: dataKey + "Total",
          rightYScaleKey: dataKey + "Total",
          rightYAxisKeys: [dataKey + "Out", dataKey + "In"],
          xAxisPadding: 12,
        });
        chart.plot(BarPlot, {
          id: dataKey + "Total",
          key: dataKey + "Total",
          legendLabel: label + " total",
          // areaGradient: ["rgba(133, 196, 255, 0.5)", "rgba(87, 211, 205, 0.5)"],
          barColor: "rgba(133, 196, 255, 0.5)",
        });
        chart.plot(LinePlot, {
          id: dataKey + "Out",
          key: dataKey + "Out",
          legendLabel: label + " out",
          lineColor: "#BA93FA",
        });
        chart.plot(LinePlot, {
          id: dataKey + "In",
          key: dataKey + "In",
          legendLabel: label + " in",
          lineColor: "#FAAE93",
        });
        chart.draw();
      }

      setInterval(navCheck(container), 200);
    },
    [data, data ? data[0].date : "", data ? data[data.length - 1].date : "", dataKey, settings, label]
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

export default ProfitsChart;
