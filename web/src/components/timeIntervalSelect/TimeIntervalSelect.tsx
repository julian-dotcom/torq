import "./interval_select.scss";
import { Fragment, useState } from "react";
import { format } from "date-fns";
import { defaultStaticRanges, defineds } from "./customRanges";
import { Menu, Transition } from "@headlessui/react";
import { CalendarLtr20Regular as IntervalIcon } from "@fluentui/react-icons";
import { DateRangePicker } from "react-date-range";
import { Popover } from "react-tiny-popover";
import { addDays } from "date-fns";

function classNames(...classes: any) {
  return classes.filter(Boolean).join(" ");
}

function RangeItem(props: any) {
  const item = props.range;
  return (
    <Menu.Item>
      {({ active }) => (
        <p
          onClick={() =>
            props.setCurrentPeriod([
              item.range().startDate,
              item.range().endDate,
              item.rangeCompare().startDate,
              item.rangeCompare().endDate,
            ])
          }
          className={classNames(
            active ? "bg-gray-100 text-gray-900" : "text-gray-700",
            "block px-4 py-2 text-sm"
          )}
        >
          {item.label}
        </p>
      )}
    </Menu.Item>
  );
}

function TimeIntervalSelect() {
  const [currentPeriod, setCurrentPeriod] = useState([
    defineds.startOfLastWeek,
    defineds.startOfToday,
    defineds.startOfLastWeekCompare,
    defineds.endOfLastWeekCompare,
  ]);

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
    <div>
      <Popover
        onClickOutside={() => setIsPopoverOpen(!isPopoverOpen)}
        containerClassName="date-range-popover"
        isOpen={isPopoverOpen}
        positions={["bottom"]}
        content={
          <div className="shadow-lg">
            <div style={{ background: "white" }}>
              <DateRangePicker
                monthDisplayFormat="MMMM yyyy"
                showDateDisplay={false}
                staticRanges={defaultStaticRanges}
                rangeColors={["#ECFAF8", "#F9FAFB"]}
                maxDate={new Date()}
                scroll={{ enabled: true }}
                months={1}
                showMonthArrow={false}
                showMonthAndYearPickers={false}
                direction="vertical"
                inputRanges={[]}
                ranges={[state.selection1]}
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
          <div className="icon">{/* <IntervalIcon /> */}</div>
          <div className="interval">
            <div className="justify-center w-full py-2 bg-white text-sm font-medium text-gray-700 focus:outline-none focus:ring-2focus:ring-offset-gray-100 focus:ring-indigo-500">
              <p className="text-base">
                {" "}
                {format(currentPeriod[0], "MMM d, yyyy")} -{" "}
                {format(currentPeriod[1], "MMM d, yyyy")}
              </p>
              <p className="text-slate-400 text-sm">
                {" "}
                {format(currentPeriod[2], "MMM d, yyyy")} -{" "}
                {format(currentPeriod[3], "MMM d, yyyy")}
              </p>
            </div>
          </div>
        </div>
      </Popover>
    </div>
  );
}
export default TimeIntervalSelect;
