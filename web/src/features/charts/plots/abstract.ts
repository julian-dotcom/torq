import ChartCanvas from "../chartCanvas";

export type basePlotConfig = {
  id: string; // The id used to fetch the Plot instance from the Chart instance
  key: string; // The id used to fetch the Plot instance from the Chart instance
};

export type drawConfig = {
  xPosition?: number;
  yPosition?: number;
  xValue?: Date;
  leftYValue?: number;
  rightYValue?: number;
  xIndex?: number;
};

export abstract class AbstractPlot {
  chart: ChartCanvas;
  config: basePlotConfig;

  constructor(chart: ChartCanvas, config: basePlotConfig) {
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
    return this.chart.config.yScale(dataPoint);
  }

  /**
   * xPoint returns the starting location for the bar on the xScale in pixels
   *
   * @param xValue the data point on the xScale that you want to convert to a pixel location on the chart.
   */
  xPoint(xValue: number | Date): number {
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
