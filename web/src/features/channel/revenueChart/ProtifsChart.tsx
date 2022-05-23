// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import ChartCanvas from "../../charts/chartCanvas";
import { BarPlot } from "../../charts/plots/bar";
import "../../charts/chart.scss";
import chartCanvas from "../../charts/chartCanvas";

type ProfitsChart = {
  data: any[];
};

function ProfitsChart({ data }: ProfitsChart) {
  let chart: ChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];

  // Check and update the chart size if the navigation changes the container size
  const navCheck: Function = (container: Selection<HTMLDivElement, {}, HTMLElement, any>): Function => {
    return () => {
      let boundingBox = container?.node()?.getBoundingClientRect();
      if (currentSize[0] !== boundingBox?.width || currentSize[1] !== boundingBox?.height) {
        chart.resizeChart();
        chart.draw();
        currentSize = [boundingBox?.width, boundingBox?.height];
      }
    };
  };

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      chart = new ChartCanvas(container, data, {
        leftYAxisKey: "revenue_out",
        xAxisPadding: 12,
      });
      chart.plot(BarPlot, { id: "bars", key: "revenue_out" });
      chart.draw();
      setInterval(navCheck(container), 200);
    },
    [data]
  );

  useEffect(() => {
    return () => {
      if (chart) {
        chart.removeResizeListener();
      }
    };
  }, [data]);

  // @ts-ignore
  return <div ref={ref} className={"chart-ref"} />;
}

export default ProfitsChart;
