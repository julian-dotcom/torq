// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "../../charts/useD3";
import React, { useEffect } from "react";
import { Selection } from "d3";
import FlowChartCanvas, { FlowData } from "../../charts/flowChartCanvas";

type FlowChart = {
  data: Array<FlowData>;
};

function FlowChart({ data }: FlowChart) {
  let flowChart: FlowChartCanvas;

  // TODO: Change this so that we can update the data without redrawing the entire chart
  const ref = useD3(
    (container: Selection<HTMLDivElement, {}, HTMLElement, any>) => {
      flowChart = new FlowChartCanvas(container, data, {});
      flowChart.draw();
    },
    [data]
  );

  useEffect(() => {
    return () => {
      if (flowChart) {
        flowChart.removeResizeListener();
      }
    };
  }, [data]);

  // @ts-ignore
  return <div ref={ref} />;
}

export default FlowChart;
