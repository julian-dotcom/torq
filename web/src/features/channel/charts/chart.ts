import * as d3 from "d3";
import { NumberValue, ScaleLinear, ScaleTime, Selection } from "d3";
import { addHours, subHours } from "date-fns";
import clone from "clone";
import eventIcons from "./eventIcons";

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

class Chart {
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

  figures: Map<string, { plot: BarPlot; interactionConfig: any }> = new Map<
    string,
    { plot: BarPlot; interactionConfig: barInputConfig }
  >();

  getFigure(xLocation: number, yLocation: number): { plot: BarPlot; interactionConfig: any } | undefined {
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
type basePlotConfig = {
  id: string; // The id used to fetch the Plot instance from the Chart instance
  key: string; // The id used to fetch the Plot instance from the Chart instance
};

type drawConfig = {
  xPosition?: number;
  yPosition?: number;
  xValue?: Date;
  leftYValue?: number;
  rightYValue?: number;
  xIndex?: number;
};

abstract class AbstractPlot {
  chart: Chart;
  config: basePlotConfig;

  constructor(chart: Chart, config: basePlotConfig) {
    this.chart = chart;

    this.config = {
      ...config,
    };
  }

  setConfig(config: basePlotConfig) {
    this.config = { ...this.config, ...config };
  }

  getYScale() {
    const yScaleMax = Math.max(
      ...this.chart.data.map((d): number => {
        return d[this.config.key];
      })
    );
    return this.chart.config.yScale.copy().domain([yScaleMax * this.chart.config.yAxisPadding, 0]);
  }

  /**
   * Used to offset the y pixel location effectively flipping the graph vertical axis,
   * which is necessary because drawing has origin top left, while a chart/graph has origin bottom left.
   */
  offset(): number {
    return (this.chart.canvas.node()?.height || 0) - this.chart.config.yScale.range()[1];
  }

  height(dataPoint: number): number {
    return this.getYScale()(dataPoint);
  }

  /**
   * xPoint returns the starting location for the bar on the xScale in pixels
   *
   * @param xValue the data point on the xScale that you want to convert to a pixel location on the chart.
   */
  xPoint(xValue: number): number {
    return this.chart.config.xScale(xValue) || 0;
  }

  /**
   * yPoint returns the starting location for the bar on the yScale in pixels
   *
   * @param yValue the data point on the yScale that you want to convert to a pixel location on the chart.
   */
  yPoint(yValue: number): number {
    return this.height(yValue) + this.offset();
  }

  abstract draw(drawConfig: drawConfig): any;
}

type barsConfig = basePlotConfig & {
  key: string; // The key used to fetch data
  barGap: number; // The gap between each bar
  barColor: string; // The color of the bar
  barHoverColor: string;
  labels?: boolean;
  textColor: string;
  textHoverColor: string;
  cornerRadius: number; // The radius of the bar
};
type barInputConfig = Partial<barsConfig> & Required<Pick<barsConfig, "id" | "key">>;

export class BarPlot extends AbstractPlot {
  config: barsConfig;

  /**
   * Plots bars on a chart canvas. To use it add it to the plots map on the Chart instance.
   *
   * @param chart - The Chart instance where BarPlot will be plotted on
   * @param config - Plot config, only required attributes are key and ID
   */
  constructor(chart: Chart, config: barInputConfig) {
    super(chart, config);

    this.config = {
      barGap: 0.1,
      barColor: "#B6DCFF",
      barHoverColor: "#9DD0FF",
      textColor: "#8198A3",
      textHoverColor: "#3A463C",
      cornerRadius: 3,
      ...config,
    };
  }

  /**
   * xPoint returns the starting location for the bar on the xScale in pixels
   *
   * @param xValue the data point on the xScale that you want to convert to a pixel location on the chart.
   */
  xPoint(xValue: number): number {
    return (this.chart.config.xScale(xValue) || 0) - this.barWidth() / 2;
  }

  barWidth(): number {
    return (this.chart.config.xScale(new Date(1, 0, 1)) || 0) - (this.chart.config.xScale(new Date(1, 0, 0)) || 0);
  }

  drawBar(context: CanvasRenderingContext2D, dataPoint: any, fillColor: string) {
    context.fillStyle = fillColor;
    context.strokeStyle = fillColor;

    // Draw the bar rectangle
    context.fillRect(
      this.xPoint(dataPoint.date) + this.config.cornerRadius / 2 + (this.barWidth() * this.config.barGap) / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.height(-dataPoint[this.config.key]) - this.config.cornerRadius
    );

    // This draws the stroke used to create rounded corners
    context.strokeRect(
      this.xPoint(dataPoint.date) + this.config.cornerRadius / 2 + (this.barWidth() * this.config.barGap) / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.height(-dataPoint[this.config.key]) - this.config.cornerRadius
    );
  }

  /**
   * Draw draws the bars on the Chart instance based on the configuration provided.
   */
  draw(drawConfig?: drawConfig) {
    this.chart.data.forEach((data, i) => {
      this.chart.context.fillStyle = this.config.barColor;
      this.chart.context.strokeStyle = this.config.barColor;
      this.chart.context.lineJoin = "round";
      this.chart.context.lineWidth = this.config.cornerRadius;

      let xIndex: number = -1;
      if (
        drawConfig?.xPosition &&
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) > data.date &&
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) < addHours(data.date, 24)
      ) {
        xIndex = i;
      }

      const hoversOverDataPoint = xIndex === i;

      let barColor = this.config.barColor;
      if (hoversOverDataPoint) {
        barColor = this.config.barHoverColor;
      }
      this.drawBar(this.chart.context, this.chart.data[i], barColor);

      // Create the interaction color and
      const interactionColor = this.chart.genColor();
      // this.chart.figures.set(interactionColor, {
      //   plot: this,
      //   interactionConfig: { hoverIndex: i },
      // });

      this.chart.interactionContext.fillStyle = interactionColor;
      this.chart.interactionContext.fillRect(
        Math.round(this.xPoint(this.chart.data[i].date)),
        Math.round(this.yPoint(this.chart.data[i][this.config.key])),
        Math.round(this.barWidth()),
        Math.round(this.height(-this.chart.data[i][this.config.key]))
      );
    });
    this.chart.data.forEach((data, i) => {
      let xIndex: number = -1;
      if (
        drawConfig?.xPosition &&
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) > data.date &&
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) < addHours(data.date, 24)
      ) {
        xIndex = i;
      }
      const hoversOverDataPoint = xIndex === i;
      let textColor = this.config.textColor;
      if (hoversOverDataPoint) {
        textColor = this.config.textHoverColor;
      }

      if (this.config.labels || hoversOverDataPoint) {
        this.chart.context.font = "12px Inter";
        this.chart.context.textAlign = "center";
        this.chart.context.textBaseline = "middle";
        this.chart.context.fillStyle = textColor;
        this.chart.context.fillText(
          d3.format(",")(this.chart.data[i][this.config.key]),
          this.xPoint(this.chart.data[i].date) + this.barWidth() / 2,
          this.yPoint(this.chart.data[i][this.config.key]) - 15 + this.config.cornerRadius / 2
        );
      }
    });
  }
}

type areaPlotConfig = basePlotConfig & {
  areaColor: string;
  areaGradient?: Array<string>[2];
  addBuffer: boolean;
  globalAlpha: number;
  labels?: boolean;
};

type areaPlotConfigInit = Partial<areaPlotConfig> & basePlotConfig;

export class AreaPlot extends AbstractPlot {
  config: areaPlotConfig;
  legend: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendTextBox: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendColorBox: Selection<HTMLDivElement, {}, HTMLElement, any>;

  constructor(chart: Chart, config: areaPlotConfigInit) {
    super(chart, config);

    this.config = {
      areaColor: "#DAEDFF",
      globalAlpha: 1,
      addBuffer: false,
      ...config,
    };

    this.legend = this.chart.legendContainer
      .append("div")
      .attr("id", `${this.config.id}`)
      .attr("style", `display: grid; grid-auto-flow: column; align-items: center; grid-column-gap: 5px;`);

    this.legendTextBox = this.legend.append("div").attr("class", "legendTextBox");

    const legendColor = this.config.areaGradient
      ? `linear-gradient(0deg, ${this.config.areaGradient[0]} 0%, ${this.config.areaGradient[1]} 100%)`
      : this.config.areaColor;

    this.legendColorBox = this.legend
      .append("div")
      .attr("class", "legendColorBox")
      .attr("style", `width: 12px; height: 12px; background: ${legendColor};`);
  }

  draw(drawConfig?: drawConfig) {
    const area = d3
      .area()
      .x((d, i): number => {
        return this.chart.config.xScale(this.chart.data[i].date) || 0;
      })
      .y0((d, i): number => {
        return this.getYScale()(this.chart.data[i][this.config.key]) || 0;
      })
      .y1((d, i): number => {
        return this.getYScale()(0) || 1;
      })
      .context(this.chart.context);

    this.chart.context.globalAlpha = this.config.globalAlpha;

    this.chart.context.fillStyle = this.config.areaColor;

    if (this.config.areaGradient) {
      let gradient = this.chart.context.createLinearGradient(0, 0, 0, this.chart.config.yScale.range()[1] || 0);
      gradient.addColorStop(0, this.config.areaGradient[1]);
      gradient.addColorStop(1, this.config.areaGradient[0]);
      this.chart.context.fillStyle = gradient;
    }

    let data = this.chart.data;

    if (data && this.config.addBuffer) {
      let lastItem = clone(data[data.length - 1]);
      let firstItem = clone(data[0]);
      lastItem.date = addHours(lastItem.date, 16);
      data.push(lastItem);
      data.unshift(firstItem);
      firstItem.date = subHours(firstItem.date, 16);
    }

    this.chart.context.beginPath();
    area(data);

    this.chart.context.fill();
    this.chart.context.globalAlpha = 1;
    if (data && this.config.addBuffer) {
      data.splice(data.length - 1, 1);
      data.splice(0, 1);
    }

    this.chart.data.forEach((d, i) => {
      if (this.config.labels) {
        this.chart.context.font = "12px Inter";
        this.chart.context.textAlign = "center";
        this.chart.context.textBaseline = "middle";
        this.chart.context.fillStyle = "#3A463C";
        this.chart.context.fillText(
          d3.format(",")(d[this.config.key]),
          this.xPoint(d.date),
          this.yPoint(d[this.config.key]) - 15
        );
      }
    });

    const legendText = drawConfig?.xIndex
      ? this.chart.data[drawConfig?.xIndex][this.config.key]
      : this.chart.data[this.chart.data.length - 1][this.config.key];
    this.legendTextBox.text(d3.format(",")(legendText));
  }
}

type linePlotConfig = basePlotConfig & {
  lineColor: string;
  globalAlpha: number;
  labels?: boolean;
};

type linePlotConfigInit = Partial<linePlotConfig> & basePlotConfig;

export class LinePlot extends AbstractPlot {
  config: linePlotConfig;
  legend: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendTextBox: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendColorBox: Selection<HTMLDivElement, {}, HTMLElement, any>;

  constructor(chart: Chart, config: linePlotConfigInit) {
    super(chart, config);

    this.config = {
      lineColor: "#85C4FF",
      globalAlpha: 1,
      ...config,
    };

    this.legend = this.chart.legendContainer
      .append("div")
      .attr("id", `${this.config.id}`)
      .attr(
        "style",
        `display: grid; grid-auto-flow: column; align-items: center; grid-column-gap: 5px; justify-content: end;`
      );

    this.legendTextBox = this.legend.append("div").attr("class", "legendTextBox");

    this.legendColorBox = this.legend
      .append("div")
      .attr("class", "legendColorBox")
      .attr("style", `width: 12px; height: 12px; background: ${this.config.lineColor};`);
  }

  draw(drawConfig?: drawConfig) {
    const line = d3
      .line()
      .x((d, i): number => {
        return this.chart.config.xScale(this.chart.data[i].date) || 0;
      })
      .y((d, i): number => {
        return this.getYScale()(this.chart.data[i][this.config.key]) || 0;
      })
      .context(this.chart.context);

    this.chart.context.globalAlpha = this.config.globalAlpha;
    this.chart.context.strokeStyle = this.config.lineColor;

    this.chart.context.beginPath();
    line(this.chart.data);

    this.chart.context.stroke();
    this.chart.context.globalAlpha = 1;

    if (this.config.labels) {
      this.chart.data.forEach((d, i) => {
        this.chart.context.font = "12px Inter";
        this.chart.context.textAlign = "center";
        this.chart.context.textBaseline = "middle";
        this.chart.context.fillStyle = "#3A463C";
        this.chart.context.fillText(
          d3.format(",")(d[this.config.key]),
          this.xPoint(d.date),
          this.yPoint(d[this.config.key]) - 15
        );
      });
    }

    const hoverIndex = drawConfig?.xIndex || this.chart.data.length - 1;
    const legendText = this.chart.data[hoverIndex][this.config.key];
    this.legendTextBox.text(d3.format(",")(legendText));
  }
}

type eventsPlotConfig = basePlotConfig;

export class EventsPlot extends AbstractPlot {
  config: eventsPlotConfig;

  lastWidth?: number;
  lastHeight?: number;

  constructor(chart: Chart, config: eventsPlotConfig) {
    super(chart, config);

    this.config = {
      ...config,
    };
  }

  draw(drawConfig?: drawConfig) {
    // Don't redraw events if nothing has changed as it is expensive to add/remove html elements
    if (this.lastWidth === this.chart.config.width && this.lastHeight === this.chart.config.height) {
      return;
    }
    this.lastWidth = this.chart.config.width;
    this.lastHeight = this.chart.config.height;

    // Clear events
    this.chart.eventsContainer.selectAll("*").remove();

    this.chart.eventsContainer
      .selectAll(".event-wrapper")
      .data(this.chart.data)
      .enter()
      .append("div")
      .attr("class", "event-wrapper")
      .attr("id", function (d: any, i) {
        return "event-" + i;
      })
      .attr("style", (d: any, i) => {
        return `position:absolute; left: ${this.xPoint(d.date)}px; bottom:5px;`;
      })
      .selectAll(".event-item")
      .data((d: any, i: number) => {
        return d.events || [];
      })
      .enter()
      .append("div")
      .attr("class", (d: any, o) => {
        return "event-item " + d.type;
      })
      .html((d: any, i) => {
        const icon = eventIcons.get(d.type) || "";
        const text = d3.format(".2s")(d.value);
        if (d.value === undefined) {
          return icon;
        } else if (d.value === 0) {
          return "0" + icon;
        } else {
          return `<span class="event-item-text">${text}</span>` + icon;
        }
      });
  }
}

export default Chart;
