// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../useD3";
import React from "react";
import { Selection } from "d3";
import Chart, { BarPlot } from "../chart";
import "./bar-chart.scss";

function BarChart({ data }: { data: any[] }) {
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      const chart = new Chart(container, data, "revenue");
      chart.plot(BarPlot, { id: "revenue", key: "revenue" });
      chart.draw();
    },
    [data.length]
  );

  // @ts-ignore
  return <div ref={ref} />;
}

export default BarChart;
