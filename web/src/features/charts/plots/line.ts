import ChartCanvas from "../chartCanvas";
import { AbstractPlot, basePlotConfig, drawConfig } from "./abstract";
import * as d3 from "d3";
import { Selection } from "d3";

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

  constructor(chart: ChartCanvas, config: linePlotConfigInit) {
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
