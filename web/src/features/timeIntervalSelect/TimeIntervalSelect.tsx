import "./interval_select.scss";
import { useEffect, useState } from "react";
import { format, startOfDay, addDays, subDays, differenceInDays } from "date-fns";
import { DateRangePicker } from "react-date-range";
import { ChevronLeft24Regular as LeftIcon, ChevronRight24Regular as RightIcon } from "@fluentui/react-icons";

import { defaultStaticRangesFn } from "./customRanges";

import Popover from "features/popover/Popover";
import classNames from "classnames";
import Button, { ColorVariant } from "components/buttons/Button";
import { useAppSelector, useAppDispatch } from "store/hooks";
import { selectTimeInterval, updateInterval } from "./timeIntervalSlice";
import { useGetSettingsQuery } from "apiSlice";

interface selection {
  startDate: Date;
  endDate: Date;
  key: string;
}

function TimeIntervalSelect(props: { className?: string }) {
  // triggers RTK Query to get settings which are intercepted in the timeIntervalSlice as an extra reducer
  useGetSettingsQuery();

  const currentPeriod = useAppSelector(selectTimeInterval);

  const defaultStaticRanges = defaultStaticRangesFn(currentPeriod.weekStartsOn);

  const [isMobile, setIsMobile] = useState(false);

  const selection1: selection = {
    startDate: new Date(currentPeriod.from),
    endDate: new Date(currentPeriod.to),
    key: "selection1",
  };

  const dispatch = useAppDispatch();

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleChange = (item: any) => {
    const interval = {
      from: item.selection1.startDate.toString(),
      to: item.selection1.endDate.toString(),
    };
    dispatch(updateInterval(interval));
  };

  const renderCustomRangeLabel = () => (
    <div onClick={() => setIsMobile(true)} className="">
      Custom
    </div>
  );

  const dateRangeClass = classNames("date-range-container", {
    "mobile-date-range": isMobile,
  });

  const buttonText = (): string => {
    if (currentPeriod.from === currentPeriod.to) {
      return `${format(new Date(currentPeriod.from), "MMM d, yyyy")}`;
    }
    return `${format(new Date(currentPeriod.from), "MMM d, yyyy")} - ${format(
      new Date(currentPeriod.to),
      "MMM d, yyyy"
    )}`;
  };

  const moveBackwardInTime = () => {
    const diff = differenceInDays(new Date(currentPeriod.to), new Date(currentPeriod.from));
    const interval = {
      from: startOfDay(subDays(new Date(currentPeriod.from), diff + 1)).toISOString(),
      to: startOfDay(subDays(new Date(currentPeriod.to), diff + 1)).toISOString(),
    };
    dispatch(updateInterval(interval));
  };

  const moveForwardInTime = () => {
    const diff = differenceInDays(new Date(currentPeriod.to), new Date(currentPeriod.from));
    const interval = {
      from: startOfDay(addDays(new Date(currentPeriod.from), diff + 1)).toISOString(),
      to: startOfDay(addDays(new Date(currentPeriod.to), diff + 1)).toISOString(),
    };
    dispatch(updateInterval(interval));
  };

  const keydownHandler = (event: KeyboardEvent) => {
    if (event.key === "ArrowLeft") {
      moveBackwardInTime();
    }
    if (event.key === "ArrowRight") {
      moveForwardInTime();
    }
  };

  useEffect(() => {
    window.addEventListener("keydown", keydownHandler);
    return () => {
      window.removeEventListener("keydown", keydownHandler);
    };
  }, [currentPeriod]);

  const popOverButton = (
    <div className={"date-range-button"}>
      <div className="time-travel-arrow" onClick={moveBackwardInTime}>
        <LeftIcon />
      </div>
      <Button buttonColor={ColorVariant.accent1} className="time-interval-wrapper">
        {buttonText()}
      </Button>
      <div className="time-travel-arrow" onClick={moveForwardInTime}>
        <RightIcon />
      </div>
    </div>
  );

  return (
    <div className={classNames(dateRangeClass, props.className)}>
      <Popover button={popOverButton} className={"no-padding right"}>
        <div className="date-range-popover-content">
          <DateRangePicker
            renderStaticRangeLabel={renderCustomRangeLabel}
            monthDisplayFormat="MMMM yyyy"
            showDateDisplay={false}
            staticRanges={[
              ...defaultStaticRanges,
              {
                label: "Custom",
                hasCustomRendering: true,
                range: () => selection1,
                isSelected() {
                  const definedRange = this.range();
                  return (
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    defaultStaticRanges.findIndex((item: any) => {
                      // Mark Custom if definedRange is not found in predefined staticRanges
                      return (
                        item.range(currentPeriod.weekStartsOn).startDate.toString() ===
                          definedRange.startDate?.toString() &&
                        item.range(currentPeriod.weekStartsOn).endDate.toString() === definedRange.endDate?.toString()
                      );
                    }) === -1
                  );
                },
              },
            ]}
            fixedHeight={false}
            rangeColors={["#ECFAF8", "#F9FAFB"]}
            maxDate={addDays(new Date(), 1)}
            minDate={addDays(new Date().setFullYear(2015, 1, 1), 0)}
            scroll={{ enabled: true, calendarHeight: 400 }}
            months={1}
            showMonthArrow={false}
            showMonthAndYearPickers={false}
            weekStartsOn={currentPeriod.weekStartsOn as 0 | 1 | 6}
            direction="vertical"
            inputRanges={[]}
            ranges={[selection1]}
            onChange={(item) => {
              handleChange(item);
            }}
          />
          <div className="close-date-range-mobile" onClick={() => setIsMobile(false)}>
            Close
          </div>
        </div>
      </Popover>
    </div>
  );
}
export default TimeIntervalSelect;
