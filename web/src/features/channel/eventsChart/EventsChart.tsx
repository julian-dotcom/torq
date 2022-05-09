// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import Chart, { AreaPlot, LinePlot, EventsPlot } from "../charts/chart";
import "./events-chart.scss";

type EventsChart = {
  data: any[];
};

function EventsChart({ data }: EventsChart) {
  let chart: Chart;

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      chart = new Chart(container, data, "revenue");
      chart.plot(AreaPlot, {
        id: "capacity_out",
        key: "capacity_out",
        areaGradient: ["#DAEDFF", "#ABE9E6"],
      });
      chart.plot(LinePlot, { id: "line", key: "revenue" });
      chart.plot(EventsPlot, { id: "events", key: "events" });
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

export default EventsChart;
