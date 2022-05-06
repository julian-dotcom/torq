// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import Chart, { BarPlot, AreaPlot } from "../chart";
import "./bar-chart.scss";

function BarChart({ data }: { data: any[] }) {
  let chart: Chart;

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      chart = new Chart(container, data, "revenue");
      chart.plot(BarPlot, { id: "bars", key: "revenue" });
      chart.plot(AreaPlot, {
        id: "area",
        key: "revenue",
        areaGradient: ["#DAEDFF", "#DDF6F5"],
      });
      chart.draw();
    },
    []
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

export default BarChart;
