import * as d3 from "d3";
import { ScaleLinear, Selection } from "d3";

type chartConfig = {
  margin: {
    top: number;
    right: number;
    bottom: number;
    left: number;
  };
  totalInbound: number;
  totalOutbound: number;
  verticalGap: number;
  barWidth: number;
  horizontalGap: number;
  height: number;
  width: number;
  inboundStroke: string;
  outboundStroke: string;
  inboundFill: string;
  outboundFill: string;
  yScale: ScaleLinear<number, number, never>;
  xScale: ScaleLinear<number, number, number | undefined>;
};

export type FlowData = {
  node: string;
  outbound: number;
  inbound: number;
};

class FlowChartCanvas {
  config: chartConfig = {
    margin: {
      top: 0,
      right: 0,
      bottom: 0,
      left: 0,
    },
    totalInbound: 0,
    totalOutbound: 0,
    verticalGap: 5,
    barWidth: 10,
    horizontalGap: 7.5,
    height: 200,
    width: 500,
    inboundStroke: "#F3F9FF",
    outboundStroke: "#ECFAF8",
    inboundFill: "#C2E2FF",
    outboundFill: "#ABE9E6",
    yScale: d3.scaleLinear([0, 200]),
    xScale: d3.scaleLinear([0, 500]),
  };
  data: Array<FlowData> = [];

  container: Selection<HTMLDivElement, {}, HTMLElement, any>;
  chartContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  canvas: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  interactionLayer: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  context: CanvasRenderingContext2D;
  interactionContext: CanvasRenderingContext2D;

  constructor(
    container: Selection<HTMLDivElement, {}, HTMLElement, any>,
    data: Array<FlowData>,
    config: Partial<chartConfig>
  ) {
    if (container == undefined) {
      throw new Error("The chart container can't be null");
    }
    this.data = data;
    this.config = { ...this.config, ...config };

    this.container = container.attr("style", "position: relative; height: 100%;").html(null);

    // Configure the chart width and height based on the container
    this.config.width = this.getWidth();
    this.config.height = this.getHeight();

    this.config.totalInbound = data.map((d) => d.inbound).reduce((partialSum, a) => partialSum + a, 0);
    this.config.totalOutbound = data.map((d) => d.outbound).reduce((partialSum, a) => partialSum + a, 0);

    this.config.xScale = d3
      .scaleLinear()
      .range([0, this.config.width - this.config.margin.right - this.config.margin.left]);

    this.config.yScale = d3
      .scaleLinear()
      .range([0, this.config.height - this.config.margin.top - this.config.margin.bottom])
      .domain([0, Math.max(this.config.totalInbound, this.config.totalOutbound) * 1.1]);

    this.chartContainer = this.container
      .append("div")
      .attr("class", "chartContainer")
      .attr("width", this.config.width - this.config.margin.left - this.config.margin.right)
      .attr("height", this.config.height - this.config.margin.top - this.config.margin.bottom)
      .attr("style", `position: absolute; left: ${this.config.margin.left}px; top: ${this.config.margin.top}px;`);

    this.canvas = this.container
      .select(".chartContainer")
      .append("canvas")
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1])
      .attr("style", `position: absolute; left: 0; top: 0px;`);

    this.interactionLayer = this.container
      .select(".chartContainer")
      .append("canvas")
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1])
      .attr(
        "style",
        "position: absolute; left: 0; top: 0px; display: none;" // display: none;
      );

    this.context = this.canvas?.node()?.getContext("2d") as CanvasRenderingContext2D;
    this.interactionContext = this.interactionLayer?.node()?.getContext("2d") as CanvasRenderingContext2D;

    this.addResizeListener();
  }

  getHeight(): number {
    return this.container?.node()?.getBoundingClientRect().height || this.config.height;
  }

  getWidth(): number {
    return this.container?.node()?.getBoundingClientRect().width || this.config.width;
  }

  addResizeListener() {
    (d3.select(window).node() as EventTarget).addEventListener("resize", (event) => {
      this.resizeChart();
    });
  }

  removeResizeListener() {
    (d3.select(window).node() as EventTarget).removeEventListener("resize", (event) => {
      this.resizeChart();
    });
  }

  clearCanvas() {
    this.context.clearRect(0, 0, this.config.xScale.range()[1], this.config.yScale.range()[1]);
  }

  resizeChart() {
    this.config.width = this.getWidth();
    this.config.xScale.range([0, this.config.width - this.config.margin.right - this.config.margin.left]);

    this.config.height = this.getHeight();

    this.config.totalInbound = this.data.map((d) => d.inbound).reduce((partialSum, a) => partialSum + a, 0);
    this.config.totalOutbound = this.data.map((d) => d.outbound).reduce((partialSum, a) => partialSum + a, 0);

    this.config.yScale
      .range([0, this.config.height - this.config.margin.top - this.config.margin.bottom])
      .domain([0, Math.max(this.config.totalInbound, this.config.totalOutbound) * 1.1]);

    this.chartContainer
      .attr("width", this.config.width - this.config.margin.left - this.config.margin.right)
      .attr("height", this.config.height - this.config.margin.top - this.config.margin.bottom);

    this.canvas.attr("width", this.config.xScale.range()[1]).attr("height", this.config.yScale.range()[1]);

    this.clearCanvas();

    this.draw();
  }

  draw() {
    let inboundSum = 0;
    let outboundSum = 0;
    let yOffset = this.config.yScale.range()[1];
    let outboundSumPosition = this.config.xScale.range()[1] / 2 - this.config.horizontalGap;
    let inboundSumPosition = this.config.xScale.range()[1] / 2 + this.config.horizontalGap;

    let line = d3
      .line()
      .x((d) => d[0])
      .y((d) => d[1])
      .curve(d3.curveBumpX)
      .context(this.context);

    this.data
      .filter((d) => d.outbound !== 0)
      .sort((a, b) => {
        return b.outbound - a.outbound;
      })
      .forEach((d, i) => {
        this.context.fillStyle = this.config.outboundFill;

        if (d.outbound !== 0) {
          // Bars representing the amount of outbound traffic per channel with a gap between each subsequent bar
          this.context.fillRect(
            0,
            yOffset - (this.config.yScale(outboundSum) + this.config.verticalGap * i),
            this.config.barWidth,
            -this.config.yScale(d.outbound)
          );

          // Bar representing the total amount of outbound traffic, same as bars above, but without the gap
          this.context.fillRect(
            outboundSumPosition,
            yOffset - this.config.yScale(outboundSum),
            this.config.barWidth,
            -this.config.yScale(d.outbound)
          );

          this.context.fill();

          this.context.beginPath();
          line([
            [
              10,
              yOffset -
                this.config.yScale(outboundSum) -
                this.config.yScale(d.outbound) / 2 -
                this.config.verticalGap * i,
            ],
            [outboundSumPosition, yOffset - this.config.yScale(outboundSum) - this.config.yScale(d.outbound) / 2],
          ]);
          this.context.lineWidth = this.config.yScale(d.outbound);
          this.context.strokeStyle = this.config.outboundStroke;
          this.context.stroke();
          this.context.beginPath();
        }

        outboundSum += d.outbound;
      });

    this.data
      .filter((d) => d.inbound !== 0)
      .sort((a, b) => {
        return b.inbound - a.inbound;
      })
      .forEach((d, i) => {
        this.context.fillStyle = this.config.inboundFill;
        if (d.inbound !== 0) {
          // Bars representing the amount of inbound traffic per channel with a gap between each subsequent bar
          this.context.fillRect(
            inboundSumPosition,
            yOffset - this.config.yScale(inboundSum),
            10,
            -this.config.yScale(d.inbound)
          );

          // Bar representing the total amount of inbound traffic, same as bars above, but without the gap
          this.context.fillRect(
            this.config.xScale.range()[1] - this.config.barWidth,
            yOffset - (this.config.yScale(inboundSum) + this.config.verticalGap * i),
            this.config.barWidth,
            -this.config.yScale(d.inbound)
          );

          this.context.beginPath();
          line([
            [
              inboundSumPosition + this.config.barWidth,
              yOffset - this.config.yScale(inboundSum) - this.config.yScale(d.inbound) / 2,
            ],
            [
              this.config.xScale.range()[1] - this.config.barWidth,
              yOffset -
                this.config.yScale(inboundSum) -
                this.config.yScale(d.inbound) / 2 -
                this.config.verticalGap * i,
            ],
          ]);
          this.context.lineWidth = this.config.yScale(d.inbound);
          this.context.strokeStyle = this.config.inboundStroke;
          this.context.stroke();
          this.context.beginPath();
        }
        inboundSum += d.inbound;
      });

    return this;
  }
}

export default FlowChartCanvas;
