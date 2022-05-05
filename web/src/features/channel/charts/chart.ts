import * as d3 from "d3";
import { NumberValue, ScaleLinear, ScaleTime, Selection } from "d3";
import { addHours, subHours } from "date-fns";

type chartConfig = {
  data: Array<any>;
  yAxisKey: string;
  margin: {
    top: number;
    right: number;
    bottom: number;
    left: number;
  };
  height: number;
  width: number;
  yScale: ScaleLinear<number, number, never>;
  xScale: ScaleTime<number, number, number | undefined>;
};

class Chart {
  config: chartConfig = {
    data: [],
    yAxisKey: "",
    margin: {
      top: 10,
      right: 20,
      bottom: 30,
      left: 50,
    },
    height: 200,
    width: 500,
    yScale: d3.scaleLinear([0, 200]),
    xScale: d3.scaleTime([0, 800]),
  };

  data: Array<any> = [];
  plots: Map<string, object> = new Map<string, object>();

  container: Selection<HTMLDivElement, {}, HTMLElement, any>;
  canvas: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  interactionLayer: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  chartContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  xAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  yAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  eventsContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  dataLabelContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  context: CanvasRenderingContext2D;
  interactionContext: CanvasRenderingContext2D;

  constructor(
    container: Selection<HTMLDivElement, {}, HTMLElement, any>,
    data: Array<any>,
    yAxisKey: string
  ) {
    if (container == undefined) {
      throw new Error("The chart container can't be null");
    }
    this.container = container;

    this.data = data;
    this.config.yAxisKey = yAxisKey;

    this.container.attr("style", "position: relative; height: 100%;");
    // Configure the chart width and height based on the container
    this.config.width =
      this.container?.node()?.getBoundingClientRect().width ||
      this.config.width;
    this.config.height =
      this.container?.node()?.getBoundingClientRect().height ||
      this.config.height;

    let start = subHours(this.data[0].date, 16);
    let end = addHours(this.data[this.data.length - 1].date, 16);

    // Creating a scale
    // The range is the number of pixels the domain will be distributed across
    // The domain is the values to be displayed on the chart
    this.config.xScale = d3
      .scaleTime()
      .range([
        0,
        this.config.width -
          this.config.margin.right -
          this.config.margin.left -
          10,
      ])
      .domain([start, end]);

    this.config.yScale = d3
      .scaleLinear()
      .range([
        0,
        this.config.height - this.config.margin.top - this.config.margin.bottom,
      ]);

    this.container.html(null);

    this.chartContainer = this.container
      .append("div")
      .attr("class", "chartContainer")
      .attr(
        "width",
        this.config.width - this.config.margin.left - this.config.margin.right
      )
      .attr(
        "height",
        this.config.height - this.config.margin.top - this.config.margin.bottom
      )
      .attr(
        "style",
        `position: absolute; left: ${this.config.margin.left}px; top: ${this.config.margin.top}px;`
      );

    this.xAxisContainer = this.container
      .append("div")
      .attr("class", "xAxisContainer");

    this.yAxisContainer = this.container
      .append("div")
      .attr("class", "yAxisContainer");

    this.eventsContainer = this.container
      .append("div")
      .attr("class", "eventsContainer");

    this.canvas = this.container
      .select(".chartContainer")
      .append("canvas")
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1])
      .attr("style", `position: absolute; left: 10px; top: 0px;`);

    this.interactionLayer = this.container
      .select(".chartContainer")
      .append("canvas")
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1])
      .attr(
        "style",
        "position: absolute; left: 10px; top: 0px; display: none;" // display: none;
      );

    this.dataLabelContainer = this.container
      .select(".chartContainer")
      .append("div")
      .attr("class", "dataLabelContainer")
      .attr("width", this.config.xScale.range()[1] + "px")
      .attr("height", this.config.yScale.range()[1] + "px")
      .attr(
        "style",
        `position: absolute; left: 10px; top: 0px; pointer-events: none;
             width: ${this.config.xScale.range()[1]}px;
             height: ${this.config.yScale.range()[1]}px;`
      )
      .html("<div>Soemthing<div>");

    this.onHover();

    this.context = this.canvas
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    this.interactionContext = this.interactionLayer
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;
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

  figures: Map<string, { plot: BarPlot; config: barInputConfig }> = new Map<
    string,
    { plot: BarPlot; config: barInputConfig }
  >();

  getFigure(
    xLocation: number,
    yLocation: number
  ): { plot: BarPlot; config: barInputConfig } | undefined {
    const colorData = this.interactionContext.getImageData(
      xLocation,
      yLocation,
      1,
      1
    ).data;

    return this.figures.get(
      "rgb(" + [colorData[0], colorData[1], colorData[2]].join(",") + ")"
    );
  }

  onHover() {
    this.canvas.on("mousemove", (event) => {
      const pointer = d3.pointer(event);
      let figure = this.getFigure(pointer[0], pointer[1]);
      if (figure !== undefined) {
        figure.plot.setConfig(figure.config);
        figure.plot.draw();
      }
    });
  }

  plot(PlotItem: typeof BarPlot, config: any) {
    this.plots.set(config.id, new PlotItem(this, config));
  }

  drawYAxis() {
    const max = Math.max(
      ...this.data.map((d): number => {
        return d[this.config.yAxisKey];
      })
    );
    this.config.yScale.domain([max, 0]);

    this.yAxisContainer
      .attr("style", `height: 100%; width: ${this.config.margin.left}px;`)
      .append("svg")
      .attr("style", `height: 100%; width: 100%;`)
      .append("g")
      .style("font-size", "12px")
      .attr(
        "transform",
        `translate(${this.config.margin.left},${this.config.margin.top})`
      )
      .call(d3.axisLeft(this.config.yScale));
  }

  drawXAxis() {
    this.xAxisContainer
      .attr(
        "style",
        `width: 100%;
               height: ${this.config.margin.bottom}px;
               position: absolute;
               bottom: 0;
               left: 0;`
      )
      .append("svg")
      .attr("style", `height: 100%; width: 100%;`)
      .append("g")
      .style("font-size", "12px")
      .attr("transform", `translate(${this.config.margin.left + 10},0)`)
      .call(
        d3
          .axisBottom(this.config.xScale)
          .tickSizeOuter(0)
          .tickFormat(
            d3.timeFormat("%d %b") as (
              domainValue: NumberValue | Date,
              index: number
            ) => string
          )
      );
  }

  draw() {
    // Draw the X and Y axis
    this.drawXAxis();
    this.drawYAxis();

    // Draw each plot on the chart
    this.plots.forEach((plot, key) => {
      (plot as BarPlot).draw();
    });

    return this;
  }
}

type barsConfig = {
  barGap: number; // The gap between each bar
  tickInterval: number; // The interval between each bar
  barColor: string; // The color of the bar
  hoverBarIndex?: number;
  cornerRadius: number; // The radius of the bar
  id: string; // The id used to fetch the BarPlot instance from the Chart instance
  key: string; // The key used to fetch data for each bar
};
type barInputConfig = Partial<barsConfig> &
  Required<Pick<barsConfig, "id" | "key">>;

export class BarPlot {
  chart: Chart;
  config: barsConfig;

  /**
   * Plots bars on a chart canvas. To use it add it to the plots map on the Chart instance.
   *
   * @param chart - The Chart instance where BarPlot will be plotted on
   * @param config - Plot config, only required attributes are key and ID
   */
  constructor(chart: Chart, config: barInputConfig) {
    this.chart = chart;

    this.config = {
      barGap: 0.1,
      barColor: "#C2E2FF",
      cornerRadius: 3,
      tickInterval:
        (this.chart.config.xScale(new Date(1, 0, 1)) || 0) -
        (this.chart.config.xScale(new Date(1, 0, 0)) || 0),
      ...config,
    };
  }

  setConfig(config: barInputConfig) {
    this.config = { ...this.config, ...config };
  }

  /**
   * xPoint returns the starting location for the bar on the xScale in pixels
   *
   * @param xValue the data point on the xScale that you want to convert to a pixel location on the chart.
   */
  xPoint(xValue: number): number {
    return (this.chart.config.xScale(xValue) || 0) - this.barWidth() / 2;
  }

  /**
   * yPoint returns the starting location for the bar on the yScale in pixels
   *
   * @param yValue the data point on the yScale that you want to convert to a pixel location on the chart.
   */
  yPoint(yValue: number): number {
    return this.barHeight(yValue) + this.offset();
  }

  /**
   * Used to offset the y pixel location effectively flipping the graph vertical axis,
   * which is necessary because drawing has origin top left, while a chart/graph has origin bottom left.
   */
  offset(): number {
    return (
      (this.chart.canvas.node()?.height || 0) -
      this.chart.config.yScale.range()[1]
    );
  }

  barWidth(): number {
    return this.config.tickInterval;
  }

  barHeight(dataPoint: number): number {
    return this.chart.config.yScale(dataPoint);
  }

  drawBar(
    context: CanvasRenderingContext2D,
    dataPoint: any,
    fillColor: string
  ) {
    context.fillStyle = fillColor;
    context.strokeStyle = fillColor;

    // Draw the bar rectangle
    context.fillRect(
      this.xPoint(dataPoint.date) + this.config.cornerRadius / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.barHeight(-dataPoint[this.config.key]) - this.config.cornerRadius
    );

    // This draws the stroke used to create rounded corners
    context.strokeRect(
      this.xPoint(dataPoint.date) + this.config.cornerRadius / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.barHeight(-dataPoint[this.config.key]) - this.config.cornerRadius
    );
  }

  /**
   * Draw draws the bars on the Chart instance based on the configuration provided.
   */
  draw() {
    for (let i: number = 0; i < this.chart.data.length; i++) {
      this.chart.context.fillStyle = this.config.barColor;
      this.chart.context.strokeStyle = this.config.barColor;
      this.chart.context.lineJoin = "round";

      this.chart.context.lineWidth = this.config.cornerRadius;

      let barColor = this.config.barColor;
      if (this.config.hoverBarIndex === i) {
        barColor = "#9DD0FF";
      }
      this.drawBar(this.chart.context, this.chart.data[i], barColor);

      // Create the interaction color and
      const interactionColor = this.chart.genColor();
      this.chart.figures.set(interactionColor, {
        plot: this,
        config: { ...this.config, hoverBarIndex: i },
      });

      this.chart.interactionContext.fillStyle = interactionColor;
      this.chart.interactionContext.fillRect(
        Math.round(this.xPoint(this.chart.data[i].date)),
        Math.round(this.yPoint(this.chart.data[i][this.config.key])),
        Math.round(this.barWidth()),
        Math.round(this.barHeight(-this.chart.data[i][this.config.key]))
      );
    }
  }
}

export default Chart;
