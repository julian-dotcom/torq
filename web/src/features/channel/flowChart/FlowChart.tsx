// https://www.pluralsight.com/guides/using-d3.js-inside-a-react-app
import { useD3 } from "features/charts/useD3";
import { useEffect } from "react";
import { Selection } from "d3";
import FlowChartCanvas from "features/charts/flowChartCanvas";
import { FlowData } from "features/channel/channelTypes";
import { useLocation, useNavigate } from "react-router-dom";
import { userEvents } from "utils/userEvents";

export const FlowChartKeyOptions = [
  { value: "amount", label: "Amount" },
  { value: "revenue", label: "Revenue" },
  { value: "count", label: "Count" },
];

type FlowChart = {
  flowKey: string;
  data: Array<FlowData>;
};

function FlowChart({ flowKey, data }: FlowChart) {
  let flowChart: FlowChartCanvas;
  let currentSize: [number | undefined, number | undefined] = [undefined, undefined];
  const navigate = useNavigate();
  const location = useLocation();
  const { track } = userEvents();

  function handleNodeClick(channelId: number) {
    const state = location?.state?.background || location || {};
    track("FlowChart Navigation", { channel_id: channelId, background: state?.pathname });
    navigate(`/analyse/inspect/${channelId}`, { state: { background: state } });
  }

  // Check and update the chart size if the navigation changes the container size
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (container: Selection<HTMLDivElement, Record<string, never>, HTMLElement, any>) => {
      const keyOut = (flowKey + "Out") as keyof Omit<
        FlowData,
        "alias" | "lndShortChannelId" | "pubKey" | "fundingTransactionHash"
      >;
      const keyIn = (flowKey + "In") as keyof Omit<
        FlowData,
        "alias" | "lndShortChannelId" | "pubKey" | "fundingTransactionHash"
      >;
      flowChart = new FlowChartCanvas(container, data, { keyOut: keyOut, keyIn: keyIn, onClick: handleNodeClick });
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
