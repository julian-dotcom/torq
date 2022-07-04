import ChartCanvas from "../chartCanvas";
import { AbstractPlot, basePlotConfig, drawConfig } from "./abstract";
import { addHours, subHours } from "date-fns";
import * as d3 from "d3";
import { Selection } from "d3";
import clone from "../../../clone";

type areaPlotConfig = basePlotConfig & {
  areaColor: string;
  areaGradient?: Array<string>[2];
  legendLabel?: string;
  addBuffer: boolean;
  labels?: boolean;
};

type areaPlotConfigInit = Partial<areaPlotConfig> & basePlotConfig;

export class AreaPlot extends AbstractPlot {
  config: areaPlotConfig;
  legend: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendTextBox: Selection<HTMLDivElement, {}, HTMLElement, any>;
  legendColorBox: Selection<HTMLDivElement, {}, HTMLElement, any>;

  constructor(chart: ChartCanvas, config: areaPlotConfigInit) {
    super(chart, config);

    this.config = {
      areaColor: "#85C4FF",
      addBuffer: false,
      ...config,
    };

    this.legend = this.chart.legendContainer
      .append("div")
      .attr("class", "legendContent")
      .attr("id", `${this.config.id}`);

    const legendColor = this.config.areaGradient
      ? `linear-gradient(0deg, ${this.config.areaGradient[0]} 0%, ${this.config.areaGradient[1]} 100%)`
      : this.config.areaColor;

    this.legendColorBox = this.legend
      .append("div")
      .attr("class", "legendColorBox")
      .attr("style", `width: 12px; height: 12px; background: ${legendColor};`);

    this.legend
      .append("div")
      .attr("class", "legendLabelBox")
      .text((this.config.legendLabel || "") + ": ");

    this.legendTextBox = this.legend.append("div").attr("class", "legendTextBox");
  }

  draw(drawConfig?: drawConfig) {
    const area = d3
      .area()
      .x((d, i): number => {
        return this.chart.config.xScale(this.chart.data[i].date) || 0;
      })
      .y0((d, i): number => {
        return this.chart.config.yScale(this.chart.data[i][this.config.key]) || 0;
      })
      .y1((d, i): number => {
        return this.chart.config.yScale(0) || 1;
      })
      .context(this.chart.context);

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

    let hoverIndex: number;
    switch (drawConfig?.xIndex) {
      case undefined:
        hoverIndex = this.chart.data.length - 1;
        break;
      case 0:
        hoverIndex = 0;
        break;
      default:
        hoverIndex = drawConfig?.xIndex || 0;
    }
    if (this.chart?.data?.length > 0) {
      const legendText = this.chart.data[hoverIndex][this.config.key];
      this.legendTextBox.text(d3.format(",")(legendText));
    }
  }
}
