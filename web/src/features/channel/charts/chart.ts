import * as d3 from "d3";
import { ScaleLinear, ScaleTime, Selection } from "d3";

type chartConfig = {
  data: Array<any>;
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

  container: Selection<HTMLDivElement, {}, HTMLElement, any>;
  canvas: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  interactionLayer: Selection<HTMLCanvasElement, {}, HTMLElement, any>;
  chartContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  xAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  yAxisContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;
  eventsContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

  constructor(container: Selection<HTMLDivElement, {}, HTMLElement, any>) {
    if (container == undefined) {
      throw new Error("The chart container can't be null");
    }
    this.container = container;

    this.container.attr("style", "position: relative; height: 100%;");
    // Configure the chart width and height based on the container
    this.config.width =
      this.container?.node()?.getBoundingClientRect().width ||
      this.config.width;
    this.config.height =
      this.container?.node()?.getBoundingClientRect().height ||
      this.config.height;

    // Creating a scale
    // The range is the number of pixels the domain will be distributed across
    // The domain is the values to be displayed on the chart
    // Here we're setting dummy content
    this.config.yScale = d3
      .scaleLinear()
      .range([
        0,
        this.config.height - this.config.margin.top - this.config.margin.bottom,
      ])
      .domain([100, 0]);

    this.config.xScale = d3
      .scaleTime()
      .range([
        0,
        this.config.width -
          this.config.margin.right -
          this.config.margin.left -
          10,
      ])
      .domain([new Date(2022, 1, 1), new Date(2022, 2, 1)]);

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

  chart() {
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
    return this;
  }
}

export default Chart;
