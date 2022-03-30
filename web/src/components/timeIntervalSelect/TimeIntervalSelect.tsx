import "./interval_select.scss";
import { useState } from "react";
import { format } from "date-fns";
import locale from 'date-fns/locale/en-US'
// import locale from 'date-fns/locale/nb'
import {
  defaultStaticRanges,
  defineds,
  getCompareRanges,
} from "./customRanges";

import { DateRangePicker } from "react-date-range";
import { Popover } from "react-tiny-popover";
import { addDays } from "date-fns";
import {useAppSelector, useAppDispatch} from "../../store/hooks";
import {selectTimeInterval, updateInterval} from "./timeIntervalSlice";

interface selection {
  startDate: Date,
  endDate: Date,
  key: string,
}

function TimeIntervalSelect() {

  const currentPeriod = useAppSelector(selectTimeInterval);

  const selection1: selection = {
      startDate: new Date(currentPeriod.from),
      endDate: new Date(currentPeriod.to),
      key: "selection1",
    }

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const dispatch = useAppDispatch()

  const HandleChange = (item: any) => {
    const interval = {
      from: item.selection1.startDate.toString(),
      to: item.selection1.endDate.toString()
    }
    dispatch(updateInterval(interval))
  };

  return (
    <div>
      <Popover
        onClickOutside={() => setIsPopoverOpen(!isPopoverOpen)}
        containerClassName="date-range-popover"
        isOpen={isPopoverOpen}
        positions={['bottom']}
        align={'end'}
        content={
          <div className="date-range-popover-content">
            <div>
              <DateRangePicker
                monthDisplayFormat="MMMM yyyy"
                showDateDisplay={false}
                staticRanges={defaultStaticRanges}
                fixedHeight={false}
                rangeColors={["#ECFAF8", "#F9FAFB"]}
                maxDate={addDays(new Date(), 0)}
                minDate={addDays((new Date().setFullYear(2015,1,1)), 0)}
                scroll={{ enabled: true, calendarHeight: 400 }}
                months={1}
                showMonthArrow={false}
                showMonthAndYearPickers={false}
                weekStartsOn={locale.options?.weekStartsOn || 0}
                direction="vertical"

                inputRanges={[]}
                ranges={[selection1]}
                onChange={(item) => {
                  HandleChange(item)
                }}
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
            <div className="">
              <p className="text-base">
                {" "}
                {format(new Date(currentPeriod.from), "MMM d, yyyy")} -{" "}
                {format(new Date(currentPeriod.to), "MMM d, yyyy")}
              </p>
              {/*<p className="text-slate-400 text-sm">*/}
              {/*  {" "}*/}
              {/*  {format(new Date(currentPeriod.compareFrom), "MMM d, yyyy")} -{" "}*/}
              {/*  {format(new Date(currentPeriod.compareTo), "MMM d, yyyy")}*/}
              {/*</p>*/}
            </div>
          </div>
        </div>
      </Popover>
    </div>
  );
}
export default TimeIntervalSelect;
