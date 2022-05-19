import ChartCanvas from "../chartCanvas";
import { AbstractPlot, basePlotConfig, drawConfig } from "./abstract";
import * as d3 from "d3";
import eventIcons from "./eventIcons";

type eventsPlotConfig = basePlotConfig;

export class EventsPlot extends AbstractPlot {
  config: eventsPlotConfig;

  lastWidth?: number;
  lastHeight?: number;

  constructor(chart: ChartCanvas, config: eventsPlotConfig) {
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
