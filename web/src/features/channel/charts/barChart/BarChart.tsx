// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../useD3";
import React from "react";
import { Selection } from "d3";
import Chart, { Bars } from "../chart";
import "./bar-chart.scss";
import { addMilliseconds, subMilliseconds } from "date-fns";

function BarChart({ data }: { data: any[] }) {
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      const chart = new Chart(container);
      const revenue = data.map((d): number => {
        return d.revenue;
      });

      const diff = data[1].date - data[0].date;
      const start = subMilliseconds(data[0].date, diff / 2);
      const end = addMilliseconds(data[data.length - 1].date, diff / 2);
      chart.config.xScale.domain([start, end]);
      chart.config.yScale.domain([Math.max(...revenue), 0]);

      chart.chart();
      let bar = new Bars(chart);
      // console.log(bar.center());
      bar.draw(data);
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
