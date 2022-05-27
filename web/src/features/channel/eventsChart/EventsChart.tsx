// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import { ChartCanvas, AreaPlot, EventsPlot, LinePlot, BarPlot } from "../../charts/charts";
import "../../charts/chart.scss";

type EventsChart = {
  data: any[];
  events: any[];
};

function EventsChart({ data, events }: EventsChart) {
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
        yScaleKey: "amount_total",
        rightYScaleKey: "amount_total",
        rightYAxisKeys: ["amount_out", "amount_in", "amount_total"],
        // leftYAxisKeys: ["amount_in"],
        // showLeftYAxisLabel: true,
        // showRightYAxisLabel: true,
        // xAxisPadding: 0,
      });
      // chart.plot(BarPlot, {
      //   id: "amount_out",
      //   key: "amount_out",
      //   areaGradient: ["#DAEDFF", "#ABE9E6"],
      // });
      chart.plot(AreaPlot, {
        id: "amount_total",
        key: "amount_total",
        legendLabel: "Amount Total",
        areaGradient: ["rgba(133, 196, 255, 0.5)", "rgba(87, 211, 205, 0.5)"],
      });
      chart.plot(LinePlot, {
        id: "amount_out",
        key: "amount_out",
        legendLabel: "Amount Out",
        lineColor: "#BA93FA",
        // rightAxis: true,
      });
      chart.plot(LinePlot, {
        id: "amount_in",
        key: "amount_in",
        legendLabel: "Amount In",
        lineColor: "#FAAE93",
        // rightAxis: true,
      });
      // chart.plot(LinePlot, { id: "amount_out", key: "amount_out" });

      chart.plot(EventsPlot, { id: "events", key: "events", events: events });
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
  }, [data, data ? data[0].date : ""]);

  // @ts-ignore
  return <div ref={ref} className={"testing"} />;
}

export default EventsChart;
