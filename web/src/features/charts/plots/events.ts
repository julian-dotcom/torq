import ChartCanvas from "../chartCanvas";
import { AbstractPlot, basePlotConfig } from "./abstract";
import * as d3 from "d3";
import eventIcons from "./eventIcons";

type eventsPlotConfig = basePlotConfig & {
  events: any[]; // eslint-disable-line @typescript-eslint/no-explicit-any
};

export class EventsPlot extends AbstractPlot {
  config: eventsPlotConfig;
  events: any[]; // eslint-disable-line @typescript-eslint/no-explicit-any
  clusteredEvents: any[] = []; // eslint-disable-line @typescript-eslint/no-explicit-any

  lastWidth?: number;
  lastHeight?: number;

  constructor(chart: ChartCanvas, config: eventsPlotConfig) {
    super(chart, config);
    this.events = config.events;
    if (config.events) {
      this.clusteredEvents = config.events.reduce((rv, x) => {
        (rv[x["date"]] = rv[x["date"]] || []).push(x);
        return rv;
      }, {});
    }
    this.config = {
      ...config,
    };
  }

  draw() {
    // Don't redraw events if nothing has changed as it is expensive to add/remove html elements
    if (this.lastWidth === this.chart.config.width && this.lastHeight === this.chart.config.height) {
      return;
    }
    this.lastWidth = this.chart.config.width;
    this.lastHeight = this.chart.config.height;

    // Clear events
    this.chart.eventsContainer.selectAll("*").remove();

    if (!this.events) {
      return;
    }

    this.chart.eventsContainer
      .selectAll(".event-wrapper")
      .data(
        this.events
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          .map((d: any, _: number) => {
            return d.date;
          })
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          .filter((value: any, index: number, self: any[]) => {
            return self.indexOf(value) === index;
          })
      )
      .enter()
      .append("div")
      .attr("class", "event-wrapper")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .attr("id", function (_: any, i) {
        return "event-" + i;
      })
      .attr("style", (d: number, _) => {
        return `position:absolute; left: ${this.xPoint(new Date(d))}px; bottom:5px;`;
      })
      .selectAll(".event-item")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .data((d: any, _: number) => {
        return this.clusteredEvents[d];
      })
      .enter()
      .append("div")
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .attr("class", (d: any, _) => {
        const outbound = d.outbound ? "outbound" : "inbound";
        return `event-item ${d.type} ${outbound}`;
      })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .html((d: any, _) => {
        const icon = eventIcons.get(d.type) || "";
        const text = d.value > 1 ? d3.format(".2s")(d.value) : d.value;
        if (d.value === null) {
          return icon;
        } else if (d.value === 0) {
          return "0" + icon;
        } else {
          return `<span class="event-item-text">${text}</span>` + icon;
        }
      });
  }
}
