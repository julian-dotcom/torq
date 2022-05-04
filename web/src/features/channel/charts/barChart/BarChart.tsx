// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../useD3";
import React from "react";
import { Selection } from "d3";
import Chart from "../chart";
import "./bar-chart.scss";

function BarChart({ data }: { data: any[] }) {
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      const chart = new Chart(container);
      chart.chart();
    },
    [data.length]
  );

  return (
    <div
      // @ts-ignore
      ref={ref}
    >
      <div className={"chartContainer"} />
      <div className={"xAxisContainer"} />
      <div className={"yAxisContainer"} />
    </div>
  );
}

export default BarChart;
