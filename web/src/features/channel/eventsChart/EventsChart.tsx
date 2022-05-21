// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import { ChartCanvas, AreaPlot, EventsPlot, LinePlot } from "../../charts/charts";
import "../../charts/chart.scss";

type EventsChart = {
  data: any[];
};

function EventsChart({ data }: EventsChart) {
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
        leftYAxisKey: "revenue",
        rightYAxisKey: "capacity_out",
        showLeftYAxisLabel: true,
        showRightYAxisLabel: true,
      });
      chart.plot(AreaPlot, {
        id: "capacity_out",
        key: "capacity_out",
        areaGradient: ["#DAEDFF", "#ABE9E6"],
      });
      chart.plot(LinePlot, { id: "line", key: "revenue" });
      chart.plot(EventsPlot, { id: "events", key: "events" });
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
  return <div ref={ref} className={"testing"} />;
}

export default EventsChart;
