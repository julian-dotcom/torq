import "./interval_select.scss";
import { useState } from "react";
import { format, startOfDay, addDays, parseISO, startOfWeek, endOfWeek, subDays, sub, add, differenceInDays } from "date-fns";
import locale from 'date-fns/locale/en-US'
import { DateRangePicker } from "react-date-range";
import {
  ChevronLeft24Regular as LeftIcon,
  ChevronRight24Regular as RightIcon,
  CalendarLtr24Regular as Calendar,
} from "@fluentui/react-icons";

import {
  defaultStaticRanges,
} from "./customRanges";

import Popover from "../popover/Popover";
import classNames from "classnames";
import DefaultButton from "../buttons/Button";
import { useAppSelector, useAppDispatch } from "../../store/hooks";
import { selectTimeInterval, updateInterval } from "./timeIntervalSlice";

interface selection {
  startDate: Date,
  endDate: Date,
  key: string,
}

function TimeIntervalSelect() {
  const currentPeriod = useAppSelector(selectTimeInterval);
  const [isMobile, setIsMobile] = useState(false)
  const [isCustom, setIsCustom] = useState(false)

  const selection1: selection = {
    startDate: new Date(currentPeriod.from),
    endDate: new Date(currentPeriod.to),
    key: "selection1",
  }

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const dispatch = useAppDispatch()

  const handleChange = (item: any) => {
    const interval = {
      from: item.selection1.startDate.toString(),
      to: item.selection1.endDate.toString()
    }
    dispatch(updateInterval(interval))
  };

  const handleMobileClick = (e: boolean) => {
    setIsMobile(e)
    setIsCustom(e)
  }

  const renderCustomRangeLabel = () => (
    //@ts-ignore
    <div onClick={() => handleMobileClick(true)} className="custom-mobile">
      Custom
    </div>
  );

  const dateRangeClass = classNames("date-range-container", {
    "mobile-date-range": isMobile
  });

  const buttonText = (): string => {
    return `${format(new Date(currentPeriod.from), "MMM d, yyyy")} - ${format(new Date(currentPeriod.to), "MMM d, yyyy")}`;
  };

  let popOverButton = <DefaultButton
    onClick={() => setIsPopoverOpen(!isPopoverOpen)}
    text={buttonText()}
    className="time-interval-wrapper"
  />

  const moveBackwardInTime = () => {
    let diff = differenceInDays(new Date(currentPeriod.to), new Date(currentPeriod.from))
    const interval = {
      from: startOfDay(subDays(new Date(currentPeriod.from), diff+1)).toISOString(),
      to: startOfDay(subDays(new Date(currentPeriod.to), diff+1)).toISOString()
    }
    dispatch(updateInterval(interval))
  }

  const moveForwardInTime = () => {
    let diff = differenceInDays(new Date(currentPeriod.to), new Date(currentPeriod.from))
    const interval = {
      from: startOfDay(addDays(new Date(currentPeriod.from), diff+1)).toISOString(),
      to: startOfDay(addDays(new Date(currentPeriod.to), diff+1)).toISOString()
    }
    dispatch(updateInterval(interval))
  }

  return (
    <div className={dateRangeClass}>
      <div className="time-travel-arrow"  onClick={moveBackwardInTime}>
        <LeftIcon />
      </div>
      <Popover button={popOverButton}>
        <div className="date-range-popover-content">
          <button className="close-date-range-mobile" onClick={() => handleMobileClick(false)}>X</button>
          <DateRangePicker
            renderStaticRangeLabel={renderCustomRangeLabel}
            monthDisplayFormat="MMMM yyyy"
            showDateDisplay={false}
            staticRanges={[...defaultStaticRanges, {
              label: 'Custom',
              hasCustomRendering: true,
              range: () => ({
                startDate: startOfDay(addDays(new Date(), -3)),
                endDate: new Date()
              }),
              isSelected() {
                return isMobile
              }
            }]}
            fixedHeight={false}
            rangeColors={["#ECFAF8", "#F9FAFB"]}
            maxDate={addDays(new Date(), 0)}
            minDate={addDays((new Date().setFullYear(2015, 1, 1)), 0)}
            scroll={{ enabled: true, calendarHeight: 400 }}
            months={1}
            showMonthArrow={false}
            showMonthAndYearPickers={false}
            weekStartsOn={locale.options?.weekStartsOn || 0}
            direction="vertical"
            inputRanges={[]}
            ranges={[selection1]}
            onChange={(item) => {
              handleChange(item)
            }}
          />
        </div>
      </Popover>
      <div className="time-travel-arrow" onClick={moveForwardInTime}>
        <RightIcon />
      </div>

    </div>
  );
}
export default TimeIntervalSelect;
