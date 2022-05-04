import * as d3 from "d3";
import { ScaleLinear, ScaleTime, Selection } from "d3";
import {
  addDays,
  addHours,
  addMilliseconds,
  subDays,
  subHours,
  subMilliseconds,
} from "date-fns";

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

    // let diff = this.data[2].date - this.data[1].date;
    let start = subHours(this.data[0].date, 16);
    let end = addHours(this.data[this.data.length - 1].date, 16);
    // const start = this.data[0].date;
    // const end = this.data[this.data.length - 1].date;

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

    this.container.select(".chartContainer").remove();
    this.container.select(".xAxisContainer").remove();
    this.container.select(".yAxisContainer").remove();
    this.container.select(".eventsContainer").remove();

    this.chartContainer = this.container
      .append("div")
      .attr("class", "chartContainer");

    this.chartContainer = this.container
      .append("div")
      .attr("class", "xAxisContainer");

    this.chartContainer = this.container
      .append("div")
      .attr("class", "yAxisContainer");

    this.chartContainer = this.container
      .append("div")
      .attr("class", "eventsContainer");

    this.chartContainer = this.container.select(".chartContainer");
    this.xAxisContainer = this.container.select(".xAxisContainer");
    this.yAxisContainer = this.container.select(".yAxisContainer");
    this.eventsContainer = this.container.select(".eventsContainer");

    this.container
      .select(".chartContainer")
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
        "position: absolute; top: 0px; left: 10px; display: none;"
      );
  }

  plot(PlotItem: typeof BarPlot, config: { id: string; key: string }) {
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
      .call(d3.axisBottom(this.config.xScale).tickSizeOuter(0));
  }

  draw() {
    this.drawXAxis();
    this.drawYAxis();

    this.plots.forEach((plot, key) => {
      (plot as BarPlot).draw();
    });

    return this;
  }
}

type figureConfig = { id: string; key: string };
type barsConfig = { id: string; key: string; barSpacing: number };

export class BarPlot {
  canvas: CanvasRenderingContext2D;
  chart: Chart;

  config: barsConfig;

  constructor(chart: Chart, config: figureConfig) {
    this.chart = chart;
    this.canvas = chart.canvas
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    const defaultConfig = {
      barSpacing: 0.1,
      id: "",
      key: "",
    };

    this.config = { ...defaultConfig, ...config };
  }

  start(dataPoint: number): number {
    return this.chart.config.xScale(dataPoint) || 0;
  }

  offset(): number {
    return (
      (this.chart.canvas.node()?.height || 0) -
      this.chart.config.yScale.range()[1]
    );
  }

  barWidth(): number {
    const ticks = this.chart.config.xScale.ticks();

    return (
      // @ts-ignore
      (this.chart.config.xScale(new Date(1, 0, 1)) -
        // @ts-ignore
        this.chart.config.xScale(new Date(1, 0, 0))) *
      (1 - this.config.barSpacing)
    );
  }

  top(dataPoint: number): number {
    return this.chart.config.yScale(dataPoint);
  }

  draw() {
    for (let i: number = 0; i < this.chart.data.length; i++) {
      this.canvas.fillStyle = "#C2E2FF";
      this.canvas.strokeStyle = "#C2E2FF";
      this.canvas.lineJoin = "round";
      const cornerRadius = 5;
      this.canvas.lineWidth = cornerRadius;

      this.canvas.fillRect(
        this.start(this.chart.data[i].date) -
          this.barWidth() / 2 +
          cornerRadius / 2,
        this.top(this.chart.data[i][this.config.key]) +
          this.offset() +
          cornerRadius / 2,
        this.barWidth() - cornerRadius,
        this.top(-this.chart.data[i][this.config.key]) - cornerRadius
      );

      this.canvas.strokeRect(
        this.start(this.chart.data[i].date) -
          this.barWidth() / 2 +
          cornerRadius / 2,
        this.top(this.chart.data[i][this.config.key]) +
          this.offset() +
          cornerRadius / 2,
        this.barWidth() - cornerRadius,
        this.top(-this.chart.data[i][this.config.key]) - cornerRadius
      );
    }
  }
}

export default Chart;
