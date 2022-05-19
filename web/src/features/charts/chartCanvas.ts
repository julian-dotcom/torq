import * as d3 from "d3";
import { NumberValue, ScaleLinear, ScaleTime, Selection } from "d3";
import { addHours, subHours } from "date-fns";
import { BarPlot } from "./plots/bar";

type chartConfig = {
  leftYAxisKey: string;
  rightYAxisKey: string;
  showLeftYAxisLabel: boolean;
  showRightYAxisLabel: boolean;
  leftYAxisFormatter: (n: number | { valueOf(): number }) => string;
  rightYAxisFormatter: (n: number | { valueOf(): number }) => string;
  xAxisPadding: number;
  yAxisPadding: number;
  margin: {
    top: number;
    right: number;
    bottom: number;
    left: number;
  };
  height: number;
  width: number;
  yScale: ScaleLinear<number, number, never>;
  rightYScale: ScaleLinear<number, number, never>;
  xScale: ScaleTime<number, number, number | undefined>;
};

class ChartCanvas {
  config: chartConfig = {
    xAxisPadding: 0,
    yAxisPadding: 1.2,
    leftYAxisKey: "",
    rightYAxisKey: "",
    showLeftYAxisLabel: false,
    showRightYAxisLabel: false,
    leftYAxisFormatter: d3.format(",.3s"),
    rightYAxisFormatter: d3.format(",.3s"),
    margin: {
      top: 10,
      right: 10,
      bottom: 30,
      left: 0,
    },
    height: 200,
    width: 500,
    yScale: d3.scaleLinear([0, 200]),
    rightYScale: d3.scaleLinear([0, 200]),
    xScale: d3.scaleTime([0, 800]),
  };

  data: Array<any> = [];
  plots: Map<string, object> = new Map<string, object>();

  container: Selection<HTMLDivElement, {}, HTMLElement, any>;
  canvas: Selection<HTMLCanvasElement, {}, HTMLElement, any>;

  interactionLayer: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  chartContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  xAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  leftYAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  rightYAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  eventsContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  leftYAxisLabelContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  rightYAxisLabelContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  context: CanvasRenderingContext2D;
  interactionContext: CanvasRenderingContext2D;

  constructor(
    container: Selection<HTMLDivElement, {}, HTMLElement, any>,
    data: Array<any>,
    config: Partial<chartConfig>
  ) {
    if (container == undefined) {
      throw new Error("The chart container can't be null");
    }
    this.data = data;
    this.config = { ...this.config, ...config };

    if (this.config.leftYAxisKey) {
      this.config.margin.left = 50;
    }
    if (this.config.rightYAxisKey) {
      this.config.margin.right = 50;
    }

    this.container = container;
    this.container.attr("style", "position: relative; height: 100%;");
    // Configure the chart width and height based on the container
    this.config.width = this.getWidth();
    this.config.height = this.getHeight();

    let start = subHours(this.data[0].date, this.config.xAxisPadding);
    let end = addHours(this.data[this.data.length - 1].date, this.config.xAxisPadding);

    // Creating a scale
    // The range is the number of pixels the domain will be distributed across
    // The domain is the values to be displayed on the chart
    this.config.xScale = d3
      .scaleTime()
      .range([0, this.config.width - this.config.margin.right - this.config.margin.left])
      .domain([start, end]);

    this.config.yScale = d3
      .scaleLinear()
      .range([0, this.config.height - this.config.margin.top - this.config.margin.bottom]);

    this.config.rightYScale = d3
      .scaleLinear()
      .range([0, this.config.height - this.config.margin.top - this.config.margin.bottom]);

    this.container.html(null);

    this.chartContainer = this.container
      .append("div")
      .attr("class", "chartContainer")
      .attr("width", this.config.width - this.config.margin.left - this.config.margin.right)
      .attr("height", this.config.height - this.config.margin.top - this.config.margin.bottom)
      .attr("style", `position: absolute; left: ${this.config.margin.left}px; top: ${this.config.margin.top}px;`);

    this.legendContainer = this.container
      .append("div")
      .attr("class", "legendContainer")
      .attr(
        "style",
        `position: absolute; right: ${this.config.margin.right}px; top: ${this.config.margin.top + 10}px;`
      );

    this.xAxisContainer = this.container
      .append("div")
      .attr("class", "xAxisContainer")
      .attr(
        "style",
        `width: 100%;
               height: ${this.config.margin.bottom}px;
               position: absolute;
               bottom: 0;
               left: 0;`
      );

    this.leftYAxisContainer = this.container
      .append("div")
      .attr("class", "leftYAxisContainer")
      // .attr("style", `position: absolute; top: 0; left: 0;  height: 100%; width: ${this.config.margin.left}px;`);
      .attr(
        "style",
        `width: ${this.config.margin.left}px;
        height: ${this.config.height - this.config.margin.top}px;
        position: absolute; left: 0; top: ${this.config.margin.top}px;`
      );
    this.leftYAxisLabelContainer = this.leftYAxisContainer
      .append("div")
      .attr("class", "leftYAxisLabelContainer")
      .attr("style", `display: none;`);

    this.rightYAxisContainer = this.container
      .append("div")
      .attr("class", "rightYAxisContainer")
      // .attr("style", `position: absolute; top: 0; right: 0; height: 100%; width: ${this.config.margin.right}px;`);
      .attr(
        "style",
        `width: ${this.config.margin.right}px;
        height: ${this.config.height - this.config.margin.top}px;
        position: absolute; right: 0; top: ${this.config.margin.top}px;`
      );

    this.rightYAxisLabelContainer = this.rightYAxisContainer
      .append("div")
      .attr("class", "rightYAxisLabelContainer")
      .attr("style", `display: none;`);

    this.eventsContainer = this.container
      .append("div")
      .attr("class", "eventsContainer")
      .attr(
        "style",
        `width: ${this.config.width - this.config.margin.left - this.config.margin.right}px;
        height: ${this.config.height - this.config.margin.top - this.config.margin.bottom}px;
        position: absolute; left: ${this.config.margin.left}px; bottom: ${this.config.margin.bottom}px;`
      );

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

    // Add event listeners like hover and window resizing
    this.addResizeListener();
    this.addHoverListener();
    this.addMouseOutListener();
  }

  tickWidth(): number {
    return (this.config.xScale(new Date(1, 0, 1)) || 0) - (this.config.xScale(new Date(1, 0, 0)) || 0);
  }

  removeResizeListener() {
    (d3.select(window).node() as EventTarget).removeEventListener("resize", (event) => {
      this.resizeChart();
    });
  }

  addResizeListener() {
    (d3.select(window).node() as EventTarget).addEventListener("resize", (event) => {
      this.resizeChart();
    });
    this.drawXAxis();
    this.drawLeftYAxis();
    this.drawRightYAxis();
  }

  resizeChart() {
    this.config.width = this.getWidth();
    this.config.height = this.getHeight();

    this.config.xScale.range([0, this.config.width - this.config.margin.right - this.config.margin.left]);
    this.config.yScale.range([0, this.config.height - this.config.margin.top - this.config.margin.bottom]);
    this.config.rightYScale.range([0, this.config.height - this.config.margin.top - this.config.margin.bottom]);

    this.chartContainer
      .attr("width", this.config.width - this.config.margin.left - this.config.margin.right)
      .attr("height", this.config.height - this.config.margin.top - this.config.margin.bottom);

    this.canvas.attr("width", this.config.xScale.range()[1]).attr("height", this.config.yScale.range()[1]);

    this.interactionLayer.attr("width", this.config.xScale.range()[1]).attr("height", this.config.yScale.range()[1]);

    this.context = this.canvas?.node()?.getContext("2d") as CanvasRenderingContext2D;
    this.interactionContext = this.interactionLayer?.node()?.getContext("2d") as CanvasRenderingContext2D;
    this.context.imageSmoothingEnabled = false;
    this.interactionContext.imageSmoothingEnabled = false;

    this.draw();
  }

  getHeight(): number {
    return this.container?.node()?.getBoundingClientRect().height || this.config.height;
  }

  getWidth(): number {
    return this.container?.node()?.getBoundingClientRect().width || this.config.width;
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
      this.nextCol += 1;
    }
    return "rgb(" + ret.join(",") + ")";
  }

  figures: Map<string, { plot: BarPlot; drawConfig: any }> = new Map<string, { plot: BarPlot; drawConfig: any }>();

  getFigure(xLocation: number, yLocation: number): { plot: BarPlot; drawConfig: any } | undefined {
    const colorData = this.interactionContext.getImageData(xLocation, yLocation, 1, 1).data;

    return this.figures.get("rgb(" + [colorData[0], colorData[1], colorData[2]].join(",") + ")");
  }

  clearCanvas() {
    this.context.clearRect(0, 0, this.config.xScale.range()[1], this.config.yScale.range()[1]);
    this.interactionContext.clearRect(0, 0, this.config.xScale.range()[1], this.config.yScale.range()[1]);
  }

  addMouseOutListener() {
    this.canvas.on("mouseleave", (event) => {
      this.clearCanvas();
      this.plots.forEach((plot: any, key: string) => {
        plot.draw({});
      });
      this.drawLeftYAxisLabel(0, 0);
      this.drawRightYAxisLabel(0, 0);
    });
  }

  addHoverListener() {
    this.canvas.on("mousemove", (event) => {
      const [xPosition, yPosition] = d3.pointer(event);
      let figure = this.getFigure(xPosition, yPosition);
      this.clearCanvas();

      let xIndex: number | undefined = undefined;
      this.data.forEach((d: any, i) => {
        if (
          addHours(this.config.xScale.invert(xPosition), 12) >= d?.date &&
          addHours(this.config.xScale.invert(xPosition), 12) <= addHours(this.data[i].date, 24)
        ) {
          xIndex = i;
        }
      });

      const leftYValue = xIndex !== undefined ? this.data[xIndex || 0][this.config.leftYAxisKey] : 0;
      const rightYValue = xIndex !== undefined ? this.data[xIndex || 0][this.config.rightYAxisKey] : 0;
      this.plots.forEach((plot: any, key: string) => {
        plot.draw({
          xPosition,
          yPosition,
          xValue: this.config.xScale.invert(xPosition),
          leftYValue: leftYValue,
          rightYValue: rightYValue,
          xIndex,
        });
      });

      if (this.config.showLeftYAxisLabel && this.config.leftYAxisKey && leftYValue) {
        this.drawLeftYAxisLabel(this.config.yScale(leftYValue), leftYValue);
        this.drawYCrosshair(this.config.xScale(this.data[xIndex || 0].date) || 0, this.config.yScale(leftYValue) || 0);
      }

      if (this.config.showRightYAxisLabel && this.config.rightYAxisKey && rightYValue) {
        this.drawRightYAxisLabel(this.config.rightYScale(rightYValue), rightYValue);
        this.drawYCrosshair(
          this.config.xScale(this.data[xIndex || 0].date) || 0,
          this.config.rightYScale(rightYValue) || 0,
          true
        );
      }

      this.drawXCrosshair(
        this.config.xScale(this.data[xIndex || 0].date) || 0,
        Math.min(this.config.yScale(leftYValue), this.config.rightYScale(rightYValue))
      );
    });
  }

  plot(PlotItem: any, config: { [key: string | number]: any } & { id: string; key: string }) {
    this.plots.set(config.id, new PlotItem(this, config));
  }

  drawLeftYAxisLabel(position: number, tickLabel: number) {
    if (position === 0) {
      this.leftYAxisLabelContainer.attr("style", `top: ${position}px; display: none;`);
      return;
    }
    this.leftYAxisLabelContainer
      .attr("style", `top: ${position}px;`)
      .text(tickLabel ? this.config.leftYAxisFormatter(tickLabel) : "");
  }

  drawRightYAxisLabel(position: number, tickLabel: number) {
    if (position === 0) {
      this.rightYAxisLabelContainer.attr("style", `top: ${position}px; display: none;`);
      return;
    }
    this.rightYAxisLabelContainer
      .attr(
        "style",
        `
        top: ${position}px;`
      )
      .text(tickLabel ? this.config.rightYAxisFormatter(tickLabel) : "");
  }

  drawLeftYAxis() {
    if (this.config.leftYAxisKey === "") {
      return;
    }
    this.leftYAxisContainer.select("svg").remove();

    const max = Math.max(
      ...this.data.map((d): number => {
        return d[this.config.leftYAxisKey];
      })
    );
    this.config.yScale.domain([max * this.config.yAxisPadding, 0]);

    this.leftYAxisContainer
      .append("svg")
      .attr("style", `height: 100%; width: 100%;`)
      .append("g")
      .style("font-size", "12px")
      .attr("transform", `translate(${this.config.margin.left},0)`)
      .call(d3.axisLeft(this.config.yScale).tickFormat(d3.format(",.2s")));
  }

  drawRightYAxis() {
    if (this.config.rightYAxisKey === "") {
      return;
    }
    this.rightYAxisContainer.select("svg").remove();

    const max = Math.max(
      ...this.data.map((d): number => {
        return d[this.config.rightYAxisKey];
      })
    );
    this.config.rightYScale.domain([max * this.config.yAxisPadding, 0]);

    this.rightYAxisContainer
      .append("svg")
      .attr("style", `height: 100%; width: 100%;`)
      .append("g")
      .style("font-size", "12px")
      .attr("transform", `translate(0,0)`)
      .call(d3.axisRight(this.config.rightYScale).tickFormat(d3.format(",.2s")));
  }

  drawXAxis() {
    this.xAxisContainer.select("svg").remove();
    this.xAxisContainer
      .append("svg")
      .attr("style", `height: 100%; width: 100%;`)
      .append("g")
      .style("font-size", "12px")
      .attr("transform", `translate(${this.config.margin.left},0)`)
      .call(
        d3
          .axisBottom(this.config.xScale)
          .tickSizeOuter(0)
          .tickFormat(d3.timeFormat("%d %b") as (domainValue: NumberValue | Date, index: number) => string)
          .ticks(
            Math.min(Math.max((this.config.width - this.config.margin.left - this.config.margin.right) / 85, 2), 20)
          )
      );
  }

  drawXCrosshair(xPosition: number, yPosition: number) {
    this.context.lineWidth = 1;
    this.context.strokeStyle = "rgba(3, 48, 72, 0.4)";
    this.context.setLineDash([5, 3]);
    this.context.beginPath();
    this.context.moveTo(xPosition, yPosition);
    this.context.lineTo(xPosition, this.config.yScale.range()[1]);
    this.context.stroke();
    // Reset the dashed line setting
    this.context.setLineDash([0, 0]);
  }

  drawYCrosshair(xPosition: number, yPosition: number, right?: boolean) {
    this.context.lineWidth = 1;
    this.context.strokeStyle = "rgba(3, 48, 72, 0.4)";
    this.context.setLineDash([5, 3]);
    this.context.beginPath();
    this.context.moveTo(0, yPosition);
    if (right) {
      this.context.moveTo(this.config.xScale.range()[1], yPosition);
    }
    this.context.lineTo(xPosition, yPosition); //this.config.yScale.range()[1]
    this.context.stroke();
    // Reset the dashed line setting
    this.context.setLineDash([0, 0]);
  }

  draw() {
    // Draw the X and Y axis
    this.drawXAxis();
    this.drawLeftYAxis();
    this.drawRightYAxis();

    // Draw each plot on the chart
    this.plots.forEach((plot, key) => {
      (plot as BarPlot).draw();
    });

    return this;
  }
}

export default ChartCanvas;
