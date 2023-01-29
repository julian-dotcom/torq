import { useEffect, useState } from "react";
import { ChevronLeft24Regular as LeftIcon, ChevronRight24Regular as RightIcon } from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import "./interval_select.scss";
import { addDays, differenceInDays, format, startOfDay, subDays } from "date-fns";
import { DateRangePicker, RangeKeyDict } from "react-date-range";

import { defaultStaticRangesFn } from "./customRanges";

import Popover from "features/popover/Popover";
import classNames from "classnames";
import Button, { ButtonPosition, ColorVariant } from "components/buttons/Button";
import { useAppDispatch, useAppSelector } from "store/hooks";
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

  const handleChange = (item: RangeKeyDict) => {
    if (item?.selection1?.startDate !== undefined && item?.selection1?.endDate !== undefined) {
      const interval = {
        from: item.selection1.startDate.toString(),
        to: item.selection1.endDate.toString(),
      };
      mixpanel.track("Time Interval Change", {
        timeIntervalCurrentFrom: currentPeriod.from,
        timeIntervalCurrentTo: currentPeriod.to,
        timeIntervalNewFrom: interval.from,
        timeIntervalNewTo: interval.to,
        timeIntervalDirection: "select",
        timeIntervalDays: differenceInDays(new Date(interval.to), new Date(interval.from)),
      });
      dispatch(updateInterval(interval));
    }
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
    mixpanel.track("Time Interval Change", {
      timeIntervalCurrentFrom: currentPeriod.from,
      timeIntervalCurrentTo: currentPeriod.to,
      timeIntervalNewFrom: interval.from,
      timeIntervalNewTo: interval.to,
      timeIntervalDirection: "backwards",
      timeIntervalDays: differenceInDays(new Date(interval.to), new Date(interval.from)),
    });
    dispatch(updateInterval(interval));
  };

  const moveForwardInTime = () => {
    const diff = differenceInDays(new Date(currentPeriod.to), new Date(currentPeriod.from));
    const interval = {
      from: startOfDay(addDays(new Date(currentPeriod.from), diff + 1)).toISOString(),
      to: startOfDay(addDays(new Date(currentPeriod.to), diff + 1)).toISOString(),
    };
    mixpanel.track("Time Interval Change", {
      timeIntervalCurrentFrom: currentPeriod.from,
      timeIntervalCurrentTo: currentPeriod.to,
      timeIntervalNewFrom: interval.from,
      timeIntervalNewTo: interval.to,
      timeIntervalDirection: "forwards",
      timeIntervalDays: differenceInDays(new Date(interval.to), new Date(interval.from)),
    });
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
      <Button
        buttonColor={ColorVariant.accent1}
        className="time-interval-wrapper"
        buttonPosition={ButtonPosition.fullWidth}
      >
        {buttonText()}
      </Button>
    </div>
  );

  return (
    <div className={classNames(dateRangeClass, props.className)}>
      <Button buttonColor={ColorVariant.accent1} icon={<LeftIcon />} onClick={moveBackwardInTime} />
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
      <Button buttonColor={ColorVariant.accent1} icon={<RightIcon />} onClick={moveForwardInTime} />
    </div>
  );
}
export default TimeIntervalSelect;
