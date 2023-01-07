import React from "react";
import * as d3 from "d3";
import styles from "features/channel/channel-page.module.scss";
import classNames from "classnames";
import { format } from "date-fns";
import eventIcons from "features/charts/plots/eventIcons";
import { ChannelEventResponse, ChannelHistoryResponse, Channel, Event } from "features/channel/channelTypes";

function fm(value: number): string | number {
  if (value > 1) {
    return d3.format(",.2s")(value);
  }
  return value;
}

function formatEventText(type: string, value: number, prev: number, outbound: boolean): string {
  const changed = value > prev ? "increased" : "decreased";
  const changeText = `${changed} from ${fm(prev)} to ${fm(value)}`;
  switch (type) {
    case "fee_rate":
      return `Fee rate ${outbound ? "outbound" : "inbound"} ${changeText}`;
    case "base_fee":
      return `Base fee ${outbound ? "outbound" : "inbound"} ${changeText}`;
    case "min_htlc":
      return `Min HTLC ${outbound ? "outbound" : "inbound"} ${changeText}`;
    case "max_htlc":
      return `Max HTLC ${outbound ? "outbound" : "inbound"} ${changeText}`;
    case "rebalanced":
      return `Rebalanced ${fm(value)} ${outbound ? "outbound" : "inbound"}`;
    case "disabled":
      return `Disabled channel ${outbound ? "outbound" : "inbound"}`;
    case "enabled":
      return `Enabled channel ${outbound ? "outbound" : "inbound"}`;
  }
  return "";
}

type eventCardType = {
  events: ChannelEventResponse;
  channels: ChannelHistoryResponse;
  selectedEvents: Map<string, boolean>;
};

function EventsCard({ events, selectedEvents, channels }: eventCardType) {
  let prev: string;
  let prevAlias: string;

  return (
    <div className={classNames(styles.card, styles.scroll)} style={{ height: "600px" }}>
      <div className={styles.eventRowsWrapper}>
        {!events?.events && <div className={styles.eventRowName}>No events</div>}
        {events?.events &&
          events.events
            .filter((d: Event) => {
              return selectedEvents.get(d.type); // selectedEventTypes
            })
            .map((event: Event, index: number) => {
              const icon = eventIcons.get(event.type);
              const newDate = prev !== (event.date as string);
              const newAlias = prevAlias !== event.channelId;
              prev = event.date;
              prevAlias = event.channelId;
              const chan: Channel =
                ((channels?.channels || []).find((c: Channel) => c.channelId === event.channelId) as Channel) || {};
              console.log("channel", chan);

              return (
                <React.Fragment key={"empty-wrapper-" + index}>
                  {newDate && (
                    <div key={"date-row" + index} className={styles.eventDateRow}>
                      {format(new Date(event.date), "yyyy-MM-dd")}
                    </div>
                  )}
                  {(newDate || newAlias) && (
                    <div key={"name-row" + index} className={styles.eventRowName}>
                      <div className={styles.channelAlias}>{chan.alias}</div>
                      <div>|</div>
                      <div className={styles.channelPoint}>{chan.shortChannelId}</div>
                    </div>
                  )}
                  <div
                    key={index}
                    className={classNames(styles.eventRow, styles[event.type], styles[event.outbound ? "" : "inbound"])}
                  >
                    <div className={styles.eventRowDetails}>
                      <div className={styles.datetime}>{format(new Date(event.datetime), "hh:mm")}</div>
                      <div className={"event-type"} dangerouslySetInnerHTML={{ __html: icon as string }} />
                      <div className={"event-type-label"}>
                        {formatEventText(event.type, event.value, event.previousValue, event.outbound)}
                      </div>
                    </div>
                  </div>
                </React.Fragment>
              );
            })}
      </div>
    </div>
  );
}
export default EventsCard;
