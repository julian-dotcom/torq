import "./interval_select.scss";
import { Fragment, useState } from "react";
import { format } from "date-fns";
import { defaultStaticRanges, defineds } from "./customRanges";
import { Menu, Transition } from "@headlessui/react";

function classNames(...classes: any) {
  return classes.filter(Boolean).join(" ");
}

function RangeItem(item: any, setCurrentPeriod: any) {
  return (
    <Menu.Item>
      {({ active }) => (
        <p
          onClick={() =>
            setCurrentPeriod([
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

  return (
    <Menu as="div" className="relative inline-block text-left ml-5">
      <div>
        <Menu.Button className="justify-center w-full py-2 bg-white text-sm font-medium text-gray-700 focus:outline-none focus:ring-2focus:ring-offset-gray-100 focus:ring-indigo-500">
          <p className="text-lg">
            {" "}
            {format(currentPeriod[0], "MMM d, yyyy")} -{" "}
            {format(currentPeriod[1], "MMM d, yyyy")}
          </p>
          <p>
            {" "}
            {format(currentPeriod[2], "MMM d, yyyy")} -{" "}
            {format(currentPeriod[3], "MMM d, yyyy")}
          </p>
        </Menu.Button>
      </div>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="origin-top-right absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
          <div className="py-1">
            {defaultStaticRanges.map((range: object) =>
              RangeItem(range, setCurrentPeriod)
            )}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}
export default TimeIntervalSelect;
