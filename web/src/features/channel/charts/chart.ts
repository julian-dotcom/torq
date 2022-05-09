import * as d3 from "d3";
import {
  NumberValue,
  ScaleLinear,
  ScaleTime,
  Selection,
  svg,
  ValueFn,
} from "d3";
import { addHours, subHours } from "date-fns";
import clone from "clone";
import AddIcon from "@fluentui/svg-icons/icons/add_20_filled.svg";

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
  legendsContainer: Selection<HTMLDivElement, {}, HTMLElement, any>;

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
    this.config.width = this.getWidth();
    this.config.height = this.getHeight();

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

    this.legendsContainer = this.container
      .append("div")
      .attr("class", "legendsContainer")
      .attr(
        "style",
        `position: absolute; left: ${this.config.margin.left + 10}px; top: ${
          this.config.margin.top + 10
        }px;`
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

    this.yAxisContainer = this.container
      .append("div")
      .attr("class", "yAxisContainer")
      .attr("style", `height: 100%; width: ${this.config.margin.left}px;`);

    this.eventsContainer = this.container
      .append("div")
      .attr("class", "eventsContainer")
      .attr(
        "style",
        `width: ${
          this.config.width - this.config.margin.left - this.config.margin.right
        }px;
        height: ${
          this.config.height -
          this.config.margin.top -
          this.config.margin.bottom
        }px;
        position: absolute; left: ${this.config.margin.left}px; bottom: ${
          this.config.margin.bottom
        }px;`
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
        "position: absolute; left: 10px; top: 0px; display: none;" // display: none;
      );

    this.context = this.canvas
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    this.interactionContext = this.interactionLayer
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    // Add event listeners like hover and window resizing
    this.addResizeListener();
    this.addHoverListener();
    this.addMouseOutListener();
  }

  removeResizeListener() {
    (d3.select(window).node() as EventTarget).removeEventListener(
      "resize",
      (event) => {
        this.resizeChart();
      }
    );
  }

  addResizeListener() {
    (d3.select(window).node() as EventTarget).addEventListener(
      "resize",
      (event) => {
        this.resizeChart();
      }
    );
  }

  resizeChart() {
    this.config.width = this.getWidth();
    this.config.height = this.getHeight();

    this.config.xScale.range([
      0,
      this.config.width -
        this.config.margin.right -
        this.config.margin.left -
        10,
    ]);

    this.config.yScale.range([
      0,
      this.config.height - this.config.margin.top - this.config.margin.bottom,
    ]);

    this.chartContainer
      .attr(
        "width",
        this.config.width - this.config.margin.left - this.config.margin.right
      )
      .attr(
        "height",
        this.config.height - this.config.margin.top - this.config.margin.bottom
      );

    this.canvas
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1]);

    this.interactionLayer
      .attr("width", this.config.xScale.range()[1])
      .attr("height", this.config.yScale.range()[1]);

    this.context = this.canvas
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    this.interactionContext = this.interactionLayer
      ?.node()
      ?.getContext("2d") as CanvasRenderingContext2D;

    this.draw();
  }

  getHeight(): number {
    return (
      this.container?.node()?.getBoundingClientRect().height ||
      this.config.height
    );
  }

  getWidth(): number {
    return (
      this.container?.node()?.getBoundingClientRect().width || this.config.width
    );
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

  getFigure(
    xLocation: number,
    yLocation: number
  ): { plot: BarPlot; interactionConfig: any } | undefined {
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

  clearCanvas() {
    this.context.clearRect(
      0,
      0,
      this.config.xScale.range()[1],
      this.config.yScale.range()[1]
    );
    this.interactionContext.clearRect(
      0,
      0,
      this.config.xScale.range()[1],
      this.config.yScale.range()[1]
    );
  }

  addMouseOutListener() {
    this.canvas.on("mouseleave", (event) => {
      this.clearCanvas();
      this.plots.forEach((plot: any, key: string) => {
        plot.draw({});
      });
    });
  }

  addHoverListener() {
    this.canvas.on("mousemove", (event) => {
      const [xPosition, yPosition] = d3.pointer(event);
      let figure = this.getFigure(xPosition, yPosition);
      this.clearCanvas();

      let xIndex: number;
      this.data.forEach((d: any, i) => {
        if (
          this.config.xScale.invert(xPosition) > d?.date &&
          this.config.xScale.invert(xPosition) < addHours(this.data[i].date, 24)
        ) {
          xIndex = i;
        }
      });

      this.plots.forEach((plot: any, key: string) => {
        plot.draw({
          xPosition,
          yPosition,
          xValue: this.config.xScale.invert(xPosition),
          yValue: this.config.yScale.invert(yPosition),
          xIndex,
        });
      });
    });
  }

  plot(
    PlotItem: any,
    config: { [key: string | number]: any } & { id: string; key: string }
  ) {
    this.plots.set(config.id, new PlotItem(this, config));
  }

  drawYAxis() {
    this.yAxisContainer.select("svg").remove();

    const max = Math.max(
      ...this.data.map((d): number => {
        return d[this.config.yAxisKey];
      })
    );
    this.config.yScale.domain([max * 1.1, 0]);

    this.yAxisContainer
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
    this.xAxisContainer.select("svg").remove();
    this.xAxisContainer
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
type basePlotConfig = {
  id: string; // The id used to fetch the Plot instance from the Chart instance
  yScale: ScaleLinear<number, number, never>;
};

type drawConfig = {
  xPosition?: number;
  yPosition?: number;
  xValue?: Date;
  yValue?: number;
  xIndex?: number;
};

abstract class AbstractPlot {
  chart: Chart;
  config: basePlotConfig;

  constructor(chart: Chart, config: { id: string }) {
    this.chart = chart;

    this.config = {
      yScale: chart.config.yScale.copy(),
      ...config,
    };
  }

  setConfig(config: basePlotConfig) {
    this.config = { ...this.config, ...config };
  }

  getYScale(chart: Chart, key: string) {
    const yScaleMax = Math.max(
      ...chart.data.map((d): number => {
        return d[key];
      })
    );
    return chart.config.yScale.copy().domain([yScaleMax * 1.1, 0]);
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

  height(dataPoint: number): number {
    return this.config.yScale(dataPoint);
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
  yScale: ScaleLinear<number, number, never>;
  barGap: number; // The gap between each bar
  barColor: string; // The color of the bar
  barHoverColor: string;
  labels?: boolean;
  textColor: string;
  textHoverColor: string;
  cornerRadius: number; // The radius of the bar
};
type barInputConfig = Partial<barsConfig> &
  Required<Pick<barsConfig, "id" | "key">>;

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
      yScale: this.getYScale(chart, config.key),
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
    return (
      (this.chart.config.xScale(new Date(1, 0, 1)) || 0) -
      (this.chart.config.xScale(new Date(1, 0, 0)) || 0)
    );
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
      this.xPoint(dataPoint.date) +
        this.config.cornerRadius / 2 +
        (this.barWidth() * this.config.barGap) / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.height(-dataPoint[this.config.key]) - this.config.cornerRadius
    );

    // This draws the stroke used to create rounded corners
    context.strokeRect(
      this.xPoint(dataPoint.date) +
        this.config.cornerRadius / 2 +
        (this.barWidth() * this.config.barGap) / 2,
      this.yPoint(dataPoint[this.config.key]) + this.config.cornerRadius / 2,
      this.barWidth() * (1 - this.config.barGap) - this.config.cornerRadius,
      this.height(-dataPoint[this.config.key]) - this.config.cornerRadius
    );
  }

  // updateLegend() {
  //   this.chart.legendsContainer.select(`#${this.config.id}`).append();
  // }

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
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) >
          data.date &&
        addHours(this.chart.config.xScale.invert(drawConfig.xPosition), 12) <
          addHours(data.date, 24)
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
          this.chart.data[i][this.config.key],
          this.xPoint(this.chart.data[i].date) + this.barWidth() / 2,
          this.yPoint(this.chart.data[i][this.config.key]) -
            15 +
            this.config.cornerRadius / 2
        );
      }

      this.chart.interactionContext.fillStyle = interactionColor;
      this.chart.interactionContext.fillRect(
        Math.round(this.xPoint(this.chart.data[i].date)),
        Math.round(this.yPoint(this.chart.data[i][this.config.key])),
        Math.round(this.barWidth()),
        Math.round(this.height(-this.chart.data[i][this.config.key]))
      );
    });
  }
}

type areaPlotConfig = basePlotConfig & {
  key: string; // The key used to fetch data
  yScale: ScaleLinear<number, number, never>;
  areaColor: string;
  areaGradient?: Array<string>[2];
  addBuffer: boolean;
  globalAlpha: number;
  labels?: boolean;
};

type areaPlotConfigInit = Partial<areaPlotConfig> &
  Required<Pick<areaPlotConfig, "id" | "key">>;

export class AreaPlot extends AbstractPlot {
  config: areaPlotConfig;

  constructor(chart: Chart, config: areaPlotConfigInit) {
    super(chart, config);

    this.config = {
      areaColor: "#DAEDFF",
      yScale: this.getYScale(chart, config.key),
      globalAlpha: 1,
      addBuffer: false,
      ...config,
    };
  }

  draw(drawConfig?: drawConfig) {
    const area = d3
      .area()
      .x((d, i): number => {
        return this.chart.config.xScale(this.chart.data[i].date) || 0;
      })
      .y0((d, i): number => {
        return this.config.yScale(this.chart.data[i][this.config.key]) || 0;
      })
      .y1((d, i): number => {
        return this.config.yScale(0) || 1;
      })
      .context(this.chart.context);

    this.chart.context.globalAlpha = this.config.globalAlpha;

    this.chart.context.fillStyle = this.config.areaColor;

    if (this.config.areaGradient) {
      let gradient = this.chart.context.createLinearGradient(
        0,
        0,
        0,
        this.config.yScale.range()[1] || 0
      );
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
      if (this.config.labels || drawConfig?.xIndex === i) {
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
  }
}

type linePlotConfig = basePlotConfig & {
  key: string; // The key used to fetch data
  yScale: ScaleLinear<number, number, never>;
  lineColor: string;
  globalAlpha: number;
  labels?: boolean;
};

type linePlotConfigInit = Partial<linePlotConfig> &
  Required<Pick<linePlotConfig, "id" | "key">>;

export class LinePlot extends AbstractPlot {
  config: linePlotConfig;

  constructor(chart: Chart, config: linePlotConfigInit) {
    super(chart, config);

    this.config = {
      lineColor: "#85C4FF",
      yScale: this.getYScale(chart, config.key),
      globalAlpha: 1,
      ...config,
    };
  }

  draw(drawConfig?: drawConfig) {
    const line = d3
      .line()
      .x((d, i): number => {
        return this.chart.config.xScale(this.chart.data[i].date) || 0;
      })
      .y((d, i): number => {
        return this.config.yScale(this.chart.data[i][this.config.key]) || 0;
      })
      .context(this.chart.context);

    this.chart.context.globalAlpha = this.config.globalAlpha;
    this.chart.context.strokeStyle = this.config.lineColor;

    this.chart.context.beginPath();
    line(this.chart.data);

    this.chart.context.stroke();
    this.chart.context.globalAlpha = 1;

    this.chart.data.forEach((d, i) => {
      if (this.config.labels || drawConfig?.xIndex === i) {
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
  }
}

type eventsPlotConfig = basePlotConfig & {};

type eventsPlotConfigInit = Partial<linePlotConfig> &
  Required<Pick<linePlotConfig, "id">>;

export class EventsPlot extends AbstractPlot {
  config: eventsPlotConfig;

  icons: Map<string, string> = new Map<string, string>([
    [
      "rebalanced_in",
      `
        <svg width="16" height="17" viewBox="0 0 16 17" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M15.5 2.5896C15.7761 2.5896 16 2.81346 16 3.0896V14.0896C16 14.3657 15.7761 14.5896 15.5 14.5896C15.2239 14.5896 15 14.3657 15 14.0896V3.0896C15 2.81346 15.2239 2.5896 15.5 2.5896ZM0 8.5896C0 8.31346 0.223858 8.0896 0.5 8.0896H11.2929L8.14645 4.94315C7.95118 4.74789 7.95118 4.43131 8.14645 4.23605C8.34171 4.04078 8.65829 4.04078 8.85355 4.23605L12.8536 8.23605C12.9015 8.28398 12.9377 8.33924 12.9621 8.39821C12.9861 8.45629 12.9996 8.5199 13 8.5866L13 8.5896L13 8.5926C12.9992 8.71956 12.9504 8.84628 12.8536 8.94315L8.85355 12.9432C8.65829 13.1384 8.34171 13.1384 8.14645 12.9432C7.95118 12.7479 7.95118 12.4313 8.14645 12.236L11.2929 9.0896H0.5C0.223858 9.0896 0 8.86574 0 8.5896Z" fill="#033048"/>
        </svg>
`,
    ],
    [
      "rebalanced_out",
      `
        <svg width="16" height="17" viewBox="0 0 16 17" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M1.5 3.05237C1.77614 3.05237 2 3.27623 2 3.55237V12.0524C2 12.3285 1.77614 12.5524 1.5 12.5524C1.22386 12.5524 1 12.3285 1 12.0524V3.55237C1 3.27623 1.22386 3.05237 1.5 3.05237ZM10.6464 3.69881C10.8417 3.50355 11.1583 3.50355 11.3536 3.69881L14.8536 7.19881C15.0488 7.39408 15.0488 7.71066 14.8536 7.90592L11.3536 11.4059C11.1583 11.6012 10.8417 11.6012 10.6464 11.4059C10.4512 11.2107 10.4512 10.8941 10.6464 10.6988L13.2929 8.05237H4.5C4.22386 8.05237 4 7.82851 4 7.55237C4 7.27623 4.22386 7.05237 4.5 7.05237H13.2929L10.6464 4.40592C10.4512 4.21066 10.4512 3.89408 10.6464 3.69881Z" fill="#033048"/>
        </svg>`,
    ],
    [
      "channel_status_disabled",
      `
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M2.5 0C1.67157 0 1 0.671573 1 1.5V14.5C1 15.3284 1.67157 16 2.5 16H4.5C5.32843 16 6 15.3284 6 14.5V1.5C6 0.671573 5.32843 0 4.5 0H2.5ZM2.5 1H4.5C4.77614 1 5 1.22386 5 1.5V14.5C5 14.7761 4.77614 15 4.5 15H2.5C2.22386 15 2 14.7761 2 14.5V1.5C2 1.22386 2.22386 1 2.5 1ZM11.5 0C10.6716 0 10 0.671573 10 1.5V14.5C10 15.3284 10.6716 16 11.5 16H13.5C14.3284 16 15 15.3284 15 14.5V1.5C15 0.671573 14.3284 0 13.5 0H11.5ZM11.5 1H13.5C13.7761 1 14 1.22386 14 1.5V14.5C14 14.7761 13.7761 15 13.5 15H11.5C11.2239 15 11 14.7761 11 14.5V1.5C11 1.22386 11.2239 1 11.5 1Z" fill="#033048"/>
      </svg>

    `,
    ],
    [
      "channel_status_enabled",
      `
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M15.2204 6.68703C16.2558 7.25661 16.2558 8.74339 15.2204 9.31298L5.2234 14.812C4.22371 15.362 3 14.6393 3 13.4991L3 2.50093C3 1.36068 4.22371 0.638047 5.2234 1.18795L15.2204 6.68703ZM14.7381 8.43766C15.0833 8.2478 15.0833 7.7522 14.7381 7.56234L4.74113 2.06327C4.4079 1.87997 4 2.12084 4 2.50093L4 13.4991C4 13.8792 4.4079 14.12 4.74114 13.9367L14.7381 8.43766Z" fill="#033048"/>
      </svg>

`,
    ],
    [
      "channel_open",
      `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<g clip-path="url(#clip0_1939_30587)">
<path d="M4.19079 0.770537C4.3211 0.314446 4.73797 0 5.21231 0H10.4614C11.1865 0 11.6986 0.710426 11.4693 1.39836L11.4668 1.40584L11.4667 1.40582L10.2053 5H12.7691C13.7155 5 14.1764 6.1436 13.5356 6.81137L13.532 6.81508L13.532 6.81506L4.85551 15.6726C4.10113 16.4551 2.79636 15.7329 3.06026 14.6773L4.22998 9.99841H2.96271C2.25687 9.99841 1.74727 9.32283 1.94118 8.64415L4.19079 0.770537ZM5.21231 1C5.18445 1 5.15996 1.01847 5.15231 1.04526L2.90271 8.91887C2.89132 8.95873 2.92125 8.99841 2.96271 8.99841H4.87037C5.02434 8.99841 5.16972 9.06935 5.26447 9.19071C5.35923 9.31207 5.39279 9.47031 5.35544 9.61968L4.03041 14.9198C4.02649 14.9355 4.02679 14.9448 4.02721 14.949C4.02765 14.9534 4.02868 14.9568 4.03027 14.9601C4.03383 14.9676 4.04314 14.9798 4.06076 14.9896C4.07838 14.9993 4.09368 15.0007 4.10191 14.9997C4.10557 14.9993 4.109 14.9984 4.11295 14.9964C4.11674 14.9945 4.12478 14.9898 4.13597 14.9782L4.13952 14.9745L4.13954 14.9745L12.8151 6.11787C12.8273 6.10484 12.8305 6.09481 12.8318 6.0864C12.8336 6.07504 12.8325 6.05898 12.8253 6.04178C12.8181 6.0246 12.8079 6.01365 12.8002 6.00817C12.7949 6.00438 12.7869 6 12.7691 6H9.49996C9.33785 6 9.1858 5.9214 9.09205 5.78915C8.9983 5.65689 8.97449 5.48739 9.02817 5.33443L10.5212 1.08022C10.5331 1.04042 10.5033 1 10.4614 1H5.21231Z" fill="#033048"/>
</g>
<defs>
<clipPath id="clip0_1939_30587">
<rect width="16" height="16" fill="white"/>
</clipPath>
</defs>
</svg>

`,
    ],
    [
      "channel_close",
      `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<g clip-path="url(#clip0_1939_30575)">
<path d="M3.27265 3.97975L0.146447 0.853553C-0.0488155 0.658291 -0.0488155 0.341709 0.146447 0.146447C0.341709 -0.0488155 0.658291 -0.0488155 0.853553 0.146447L15.8536 15.1464C16.0488 15.3417 16.0488 15.6583 15.8536 15.8536C15.6583 16.0488 15.3417 16.0488 15.1464 15.8536L9.8577 10.5648L4.85429 15.6726C4.09991 16.4551 2.79514 15.7329 3.05904 14.6773L4.22876 9.99841H2.96148C2.25565 9.99841 1.74605 9.32283 1.93996 8.64415L3.27265 3.97975ZM9.15056 9.85766L4.08155 4.78865L2.90148 8.91887C2.8901 8.95873 2.92003 8.99841 2.96148 8.99841H4.86915C5.02312 8.99841 5.1685 9.06935 5.26325 9.19071C5.35801 9.31207 5.39156 9.47031 5.35422 9.61968L4.02919 14.9198C4.02527 14.9355 4.02557 14.9448 4.02599 14.949C4.02643 14.9534 4.02746 14.9568 4.02904 14.9601C4.03261 14.9676 4.04192 14.9798 4.05954 14.9896C4.07716 14.9993 4.09246 15.0007 4.10069 14.9997C4.10435 14.9993 4.10778 14.9984 4.11173 14.9964C4.11552 14.9945 4.12356 14.9898 4.13475 14.9782L4.13832 14.9745L9.15056 9.85766ZM12.8139 6.11787L10.5501 8.42896L11.2572 9.13611L13.5308 6.81506L13.5344 6.81137C14.1752 6.1436 13.7143 5 12.7678 5H10.2041L11.4655 1.40582L11.468 1.39836C11.6974 0.710426 11.1853 0 10.4602 0H5.21109C4.73675 0 4.31988 0.314446 4.18956 0.770537L3.90113 1.78004L4.71004 2.58894L5.15109 1.04526C5.15874 1.01847 5.18323 1 5.21109 1H10.4602C10.5021 1 10.5319 1.04042 10.52 1.08022L9.02695 5.33443C8.97327 5.48739 8.99708 5.65689 9.09083 5.78915C9.18458 5.9214 9.33663 6 9.49874 6H12.7678C12.7857 6 12.7936 6.00438 12.799 6.00817C12.8067 6.01365 12.8168 6.0246 12.824 6.04178C12.8313 6.05898 12.8324 6.07504 12.8306 6.0864C12.8293 6.09481 12.8261 6.10484 12.8139 6.11787Z" fill="#033048"/>
</g>
<defs>
<clipPath id="clip0_1939_30575">
<rect width="16" height="16" fill="white"/>
</clipPath>
</defs>
</svg>

`,
    ],
    [
      "channel_force_close",
      `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M8 0C12.4183 0 16 3.58172 16 8C16 12.4183 12.4183 16 8 16C3.58172 16 0 12.4183 0 8C0 3.58172 3.58172 0 8 0ZM8 1C4.13401 1 1 4.13401 1 8C1 11.866 4.13401 15 8 15C11.866 15 15 11.866 15 8C15 4.13401 11.866 1 8 1ZM8 10.5C8.41421 10.5 8.75 10.8358 8.75 11.25C8.75 11.6642 8.41421 12 8 12C7.58579 12 7.25 11.6642 7.25 11.25C7.25 10.8358 7.58579 10.5 8 10.5ZM8 4C8.24546 4 8.44961 4.17688 8.49194 4.41012L8.5 4.5V9C8.5 9.27614 8.27614 9.5 8 9.5C7.75454 9.5 7.55039 9.32312 7.50806 9.08988L7.5 9V4.5C7.5 4.22386 7.72386 4 8 4Z" fill="#033048"/>
</svg>
`,
    ],
    [
      "fee_rate",
      `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M5 7C5 5.89543 5.89543 5 7 5C8.10457 5 9 5.89543 9 7C9 8.10457 8.10457 9 7 9C5.89543 9 5 8.10457 5 7ZM7 6C6.44772 6 6 6.44772 6 7C6 7.55228 6.44772 8 7 8C7.55228 8 8 7.55228 8 7C8 6.44772 7.55228 6 7 6ZM1.5 2C0.671573 2 0 2.67157 0 3.5V10.5C0 11.3284 0.671573 12 1.5 12H12.5C13.3284 12 14 11.3284 14 10.5V3.5C14 2.67157 13.3284 2 12.5 2H1.5ZM1 3.5C1 3.22386 1.22386 3 1.5 3H3V4C3 4.55228 2.55228 5 2 5L1 5V3.5ZM1 6L2 6C3.10457 6 4 5.10457 4 4V3H10V4C10 5.10457 10.8954 6 12 6L13 6V8H12C10.8954 8 10 8.89543 10 10V11H4V10C4 8.89543 3.10457 8 2 8H1V6ZM11 3H12.5C12.7761 3 13 3.22386 13 3.5V5L12 5C11.4477 5 11 4.55228 11 4V3ZM13 9V10.5C13 10.7761 12.7761 11 12.5 11H11V10C11 9.44772 11.4477 9 12 9H13ZM3 11H1.5C1.22386 11 1 10.7761 1 10.5V9H2C2.55228 9 3 9.44772 3 10V11ZM15.0001 10.5C15.0001 11.8807 13.8808 13 12.5001 13H2.08545C2.29137 13.5826 2.84699 14 3.5001 14H12.5001C14.4331 14 16.0001 12.433 16.0001 10.5V5.49997C16.0001 4.84686 15.5827 4.29125 15.0001 4.08533V10.5Z" fill="#033048"/>
      </svg>
      `,
    ],
    [
      "base_fee",
      `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M5 7C5 5.89543 5.89543 5 7 5C8.10457 5 9 5.89543 9 7C9 8.10457 8.10457 9 7 9C5.89543 9 5 8.10457 5 7ZM7 6C6.44772 6 6 6.44772 6 7C6 7.55228 6.44772 8 7 8C7.55228 8 8 7.55228 8 7C8 6.44772 7.55228 6 7 6ZM1.5 2C0.671573 2 0 2.67157 0 3.5V10.5C0 11.3284 0.671573 12 1.5 12H12.5C13.3284 12 14 11.3284 14 10.5V3.5C14 2.67157 13.3284 2 12.5 2H1.5ZM1 3.5C1 3.22386 1.22386 3 1.5 3H3V4C3 4.55228 2.55228 5 2 5L1 5V3.5ZM1 6L2 6C3.10457 6 4 5.10457 4 4V3H10V4C10 5.10457 10.8954 6 12 6L13 6V8H12C10.8954 8 10 8.89543 10 10V11H4V10C4 8.89543 3.10457 8 2 8H1V6ZM11 3H12.5C12.7761 3 13 3.22386 13 3.5V5L12 5C11.4477 5 11 4.55228 11 4V3ZM13 9V10.5C13 10.7761 12.7761 11 12.5 11H11V10C11 9.44772 11.4477 9 12 9H13ZM3 11H1.5C1.22386 11 1 10.7761 1 10.5V9H2C2.55228 9 3 9.44772 3 10V11ZM15.0001 10.5C15.0001 11.8807 13.8808 13 12.5001 13H2.08545C2.29137 13.5826 2.84699 14 3.5001 14H12.5001C14.4331 14 16.0001 12.433 16.0001 10.5V5.49997C16.0001 4.84686 15.5827 4.29125 15.0001 4.08533V10.5Z" fill="#033048"/>
      </svg>
      `,
    ],
  ]);

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
    if (
      this.lastWidth === this.chart.config.width &&
      this.lastHeight === this.chart.config.height
    ) {
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
        return `position:absolute; left: ${
          this.xPoint(d.date) + 10
        }px; bottom:5px;`;
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
        const icon = this.icons.get(d.type) || "";
        if (d.value === undefined) {
          return icon;
        } else if (d.value === 0) {
          return "0" + icon;
        } else {
          return d3.format(".2s")(d.value) + icon;
        }
      });
  }
}

export default Chart;
