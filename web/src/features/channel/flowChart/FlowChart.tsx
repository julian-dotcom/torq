// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "features/charts/useD3";
import { useEffect } from "react";
import { Selection } from "d3";
import FlowChartCanvas from "features/charts/flowChartCanvas";
import { FlowData } from "features/channel/channelTypes";
import { useAppSelector } from "store/hooks";
import { selectFlowKeys } from "../channelSlice";

type FlowChart = {
  data: Array<FlowData>;
};

function FlowChart({ data }: FlowChart) {
  let flowChart: FlowChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];
  const flowKey = useAppSelector(selectFlowKeys);

  // Check and update the chart size if the navigation changes the container size
  const navCheck = (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
    return () => {
      const boundingBox = container?.node()?.getBoundingClientRect();
      if (currentSize[0] !== boundingBox?.width || currentSize[1] !== boundingBox?.height) {
        flowChart.resizeChart();
        flowChart.draw();
        currentSize = [boundingBox?.width, boundingBox?.height];
      }
    };
  };

  const ref = useD3(
    (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
      const keyOut = (flowKey.value + "Out") as keyof Omit<
        FlowData,
        "alias" | "lndShortChannelId" | "pubKey" | "lndChannelPoint"
      >;
      const keyIn = (flowKey.value + "In") as keyof Omit<FlowData, "alias" | "lndShortChannelId" | "pubKey" | "lndChannelPoint">;
      flowChart = new FlowChartCanvas(container, data, { keyOut: keyOut, keyIn: keyIn });
      flowChart.draw();
      setInterval(navCheck(container), 200);
    },
    [data, flowKey]
  );

  useEffect(() => {
    return () => {
      if (flowChart) {
        flowChart.removeResizeListener();
      }
    };
  }, [data, flowKey]);

  return <div ref={ref} />;
}

export default FlowChart;
