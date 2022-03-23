import "./interval_select.scss";

import { useState } from "react";
import { addDays } from "date-fns";

import { jsx, css } from "@emotion/react";

import { CalendarLtr20Regular as IntervalIcon } from "@fluentui/react-icons";
import { DateRangePicker } from "react-date-range";
import { Popover } from "react-tiny-popover";

import { defaultStaticRanges } from "./customRanges";

function TimeIntervalSelect() {
  const [state, setState] = useState({
    selection1: {
      startDate: addDays(new Date(), -7),
      endDate: new Date(),
      key: "selection1",
    },
    selection2: {
      startDate: addDays(new Date(), -15),
      endDate: addDays(new Date(), -8),
      key: "selection2",
    },
  });

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  return (
    <Popover
      onClickOutside={() => setIsPopoverOpen(!isPopoverOpen)}
      containerClassName="date-range-popover"
      isOpen={isPopoverOpen}
      positions={["bottom"]}
      content={
        <div className="shadow-lg ">
          <div style={{ background: "white" }}>
            <DateRangePicker
              staticRanges={defaultStaticRanges}
              rangeColors={["#ECFAF8", "#F9FAFB"]}
              maxDate={new Date()}
              scroll={{ enabled: true }}
              months={1}
              showMonthArrow={false}
              showMonthAndYearPickers={false}
              direction="vertical"
              inputRanges={[]}
              ranges={[state.selection1, state.selection2]}
              onChange={(item) => setState({ ...state, ...item })}
            />
          </div>
        </div>
      }
    >
      <div
        className="time-interval-wrapper"
        onClick={() => setIsPopoverOpen(!isPopoverOpen)}
      >
        <div className="icon">
          <IntervalIcon />
        </div>
        <div className="interval">
          <div className="current">3rd - 31st March</div>
          <div className="previous">1st - 28th February</div>
        </div>
      </div>
    </Popover>
  );
}
export default TimeIntervalSelect;
