// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import ChartCanvas from "../../charts/chartCanvas";
import { BarPlot } from "../../charts/plots/bar";
import "../../charts/chart.scss";

type ProfitsChart = {
  data: any[];
};

function ProfitsChart({ data }: ProfitsChart) {
  let chart: ChartCanvas;

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      chart = new ChartCanvas(container, data, { leftYAxisKey: "revenue", xAxisPadding: 12 });
      chart.plot(BarPlot, { id: "bars", key: "revenue" });
      chart.draw();
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
  return <div ref={ref} className={"testing"} />;
}

export default ProfitsChart;
