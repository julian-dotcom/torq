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

  mouseOver?: { index: number; outbound: boolean };

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
    this.context.imageSmoothingEnabled = false;
    this.interactionContext.imageSmoothingEnabled = false;
    this.addResizeListener();
    this.addHoverListener();
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

  figures: Map<string, { index: number; outbound: boolean }> = new Map<string, { index: number; outbound: boolean }>();

  getFigure(xLocation: number, yLocation: number): { index: number; outbound: boolean } | undefined {
    const colorData = this.interactionContext.getImageData(xLocation, yLocation, 1, 1).data;

    return this.figures.get("rgb(" + [colorData[0], colorData[1], colorData[2]].join(",") + ")");
  }

  addHoverListener() {
    this.canvas.on("mousemove", (event) => {
      const [xPosition, yPosition] = d3.pointer(event);
      let figure = this.getFigure(xPosition, yPosition);

      if (figure) {
        console.log(figure);
        this.mouseOver = figure;
        this.clearCanvas();
        this.draw();
      } else if (figure !== this.mouseOver) {
        this.mouseOver = undefined;
        this.clearCanvas();
        this.draw();
      }
    });
  }

  drawOutboundBars(
    context: CanvasRenderingContext2D,
    value: number,
    outboundSum: number,
    outboundSumPosition: number,
    yOffset: number,
    index: number
  ) {
    // Bars representing the amount of outbound traffic per channel with a gap between each subsequent bar
    context.fillRect(
      0,
      yOffset - (this.config.yScale(outboundSum) + this.config.verticalGap * index),
      this.config.barWidth,
      -this.config.yScale(value) // d.outbound
    );

    // Bar representing the total amount of outbound traffic, same as bars above, but without the gap
    context.fillRect(
      outboundSumPosition,
      yOffset - this.config.yScale(outboundSum),
      this.config.barWidth,
      -this.config.yScale(value)
    );
    context.fill();
  }

  drawOutboundConnectingLines(
    context: CanvasRenderingContext2D,
    value: number,
    outboundSum: number,
    outboundSumPosition: number,
    yOffset: number,
    index: number
  ) {
    let line = d3
      .line()
      .x((d) => d[0])
      .y((d) => d[1])
      .curve(d3.curveBumpX)
      .context(context);

    context.beginPath();
    line([
      [10, yOffset - this.config.yScale(outboundSum) - this.config.yScale(value) / 2 - this.config.verticalGap * index],
      [outboundSumPosition, yOffset - this.config.yScale(outboundSum) - this.config.yScale(value) / 2],
    ]);
    context.lineWidth = this.config.yScale(value);
    context.stroke();
    context.beginPath();
  }

  drawInboundBars(
    context: CanvasRenderingContext2D,
    value: number,
    inboundSum: number,
    inboundSumPosition: number,
    yOffset: number,
    index: number
  ) {
    // Bars representing the amount of inbound traffic per channel with a gap between each subsequent bar
    context.fillRect(inboundSumPosition, yOffset - this.config.yScale(inboundSum), 10, -this.config.yScale(value));

    // Bar representing the total amount of inbound traffic, same as bars above, but without the gap
    context.fillRect(
      this.config.xScale.range()[1] - this.config.barWidth,
      yOffset - (this.config.yScale(inboundSum) + this.config.verticalGap * index),
      this.config.barWidth,
      -this.config.yScale(value)
    );

    context.fill();
  }

  drawInboundConnectingLines(
    context: CanvasRenderingContext2D,
    value: number,
    inboundSum: number,
    inboundSumPosition: number,
    yOffset: number,
    index: number
  ) {
    let line = d3
      .line()
      .x((d) => d[0])
      .y((d) => d[1])
      .curve(d3.curveBumpX)
      .context(context);
    context.beginPath();
    line([
      [
        inboundSumPosition + this.config.barWidth,
        yOffset - this.config.yScale(inboundSum) - this.config.yScale(value) / 2,
      ],
      [
        this.config.xScale.range()[1] - this.config.barWidth,
        yOffset - this.config.yScale(inboundSum) - this.config.yScale(value) / 2 - this.config.verticalGap * index,
      ],
    ]);
    context.lineWidth = this.config.yScale(value);
    context.stroke();
    context.beginPath();
  }

  /**
   * nextCol keeps track of the next unique color used to identify figures (drawn objects) on the canvas.
   */
  nextCol: number = 1;

  /**
   * @remarks concept taken from https://www.freecodecamp.org/news/d3-and-canvas-in-3-steps-8505c8b27444/
   */
  genColor() {
    let ret = [];
    if (this.nextCol < 16777215) {
      ret.push(this.nextCol & 0xff);
      ret.push((this.nextCol & 0xff00) >> 8);
      ret.push((this.nextCol & 0xff0000) >> 16);
      // Increase by 10 because the drawn figure changes color when it partially touches a pixel
      // when you increase by 10, the drawn color is different enough to prevent confusion between figures
      this.nextCol += 10;
    }
    return "rgb(" + ret.join(",") + ")";
  }

  draw() {
    // Clear the interaction colours
    this.figures.clear();

    let inboundSum = 0;
    let outboundSum = 0;
    let yOffset = this.config.yScale.range()[1];
    let outboundSumPosition = this.config.xScale.range()[1] / 2 - this.config.horizontalGap;
    let inboundSumPosition = this.config.xScale.range()[1] / 2 + this.config.horizontalGap;

    this.data
      .filter((d) => d.outbound !== 0)
      .sort((a, b) => {
        return b.outbound - a.outbound;
      })
      .forEach((d, i) => {
        this.context.fillStyle = this.config.outboundFill;
        this.context.strokeStyle = this.config.outboundStroke;

        if (this.mouseOver?.index === i && this.mouseOver?.outbound === true) {
          this.context.fillStyle = "#57D3CD";
          this.context.strokeStyle = "#DDF6F5";
        }

        this.drawOutboundBars(this.context, d.outbound, outboundSum, outboundSumPosition, yOffset, i);
        this.drawOutboundConnectingLines(this.context, d.outbound, outboundSum, outboundSumPosition, yOffset, i);

        // Draw the interaction context
        const interactionColor = this.genColor();
        this.figures.set(interactionColor, { index: i, outbound: true });
        this.interactionContext.fillStyle = interactionColor;
        this.interactionContext.strokeStyle = interactionColor;
        this.drawOutboundBars(this.interactionContext, d.outbound, outboundSum, outboundSumPosition, yOffset, i);
        this.drawOutboundConnectingLines(
          this.interactionContext,
          d.outbound,
          outboundSum,
          outboundSumPosition,
          yOffset,
          i
        );

        outboundSum += d.outbound;
      });

    this.data
      .filter((d) => d.inbound !== 0)
      .sort((a, b) => {
        return b.inbound - a.inbound;
      })
      .forEach((d, i) => {
        this.context.fillStyle = this.config.inboundFill;
        this.context.strokeStyle = this.config.inboundStroke;
        if (this.mouseOver?.index === i && this.mouseOver?.outbound === false) {
          this.context.fillStyle = "#AAD6FF";
          this.context.strokeStyle = "#E7F3FF";
        }
        this.drawInboundBars(this.context, d.inbound, inboundSum, inboundSumPosition, yOffset, i);
        this.drawInboundConnectingLines(this.context, d.inbound, inboundSum, inboundSumPosition, yOffset, i);

        // Draw the interaction context
        let interactionColor = this.genColor();
        this.figures.set(interactionColor, { index: i, outbound: false });
        this.interactionContext.fillStyle = interactionColor;
        this.interactionContext.strokeStyle = interactionColor;
        this.drawInboundBars(this.interactionContext, d.inbound, inboundSum, inboundSumPosition, yOffset, i);
        this.drawInboundConnectingLines(this.interactionContext, d.inbound, inboundSum, inboundSumPosition, yOffset, i);
        inboundSum += d.inbound;
      });

    return this;
  }
}

export default FlowChartCanvas;
