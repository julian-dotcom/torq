import {
  addDays,
  endOfDay,
  startOfDay,
  startOfMonth,
  endOfMonth,
  addMonths,
  startOfWeek,
  endOfWeek,
  isSameDay,
  startOfYear,
  differenceInCalendarDays,
  addYears,
  subDays,
} from "date-fns";

// defineds: NOUN, pronounced: dɪˈfɪnedz, "The collective group of defined date ranges"
export const defineds = {
  startOfToday: startOfDay(new Date()),
  endOfToday: endOfDay(new Date()),

  startOfYesterday: startOfDay(addDays(new Date(), -1)),
  endOfYesterday: endOfDay(addDays(new Date(), -1)),
  startOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),
  endOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),

  startOfLast7Days: startOfDay(subDays(new Date(), 6)),
  endOfLast7Days: startOfDay(new Date()),
  startOfLast7DaysCompare: startOfDay(addDays(new Date(), -8)),
  endOfLast7DaysCompare: startOfDay(addDays(new Date(), -15)),

  startOfMonth: startOfMonth(new Date()),
  endOfMonth: endOfMonth(new Date()),
  startOfLastMonth: startOfMonth(addMonths(new Date(), -1)),
  endOfLastMonth: endOfMonth(addMonths(new Date(), -1)),

  last28Days: startOfDay(addDays(new Date(), -28)),
  startOfLast28DaysCompare: startOfDay(addDays(new Date(), -57)),
  endOfLast28DaysCompare: startOfDay(addDays(new Date(), -28)),

  last30Days: startOfDay(addDays(new Date(), -30)),

  startOfLast365Days: startOfDay(addYears(new Date(), -1)),

  startOfCurrentYear: startOfYear(new Date()),
  startYTDCompare: addYears(startOfYear(new Date()), -1),
  endOfYTDCompare: addYears(new Date(), -1),

  startOfLast365DaysCompare: startOfDay(addYears(new Date(), -2)),
  endOfLast365DaysCompare: startOfDay(addYears(addDays(new Date(), -1), -1)),

  startOfTime: startOfDay(new Date(2014, 1, 1, 0, 0, 0)),
};

export const getCompareRanges = (startDate, endDate) => {
  const daysDifference = differenceInCalendarDays(endDate, startDate);
  let compareRange = [addDays(startOfDay(startDate), -1), addDays(startOfDay(startDate), -(daysDifference + 1))];

  return compareRange;
};

const staticRangeHandler = {
  range: {},
  isSelected(range) {
    const definedRange = this.range();
    return isSameDay(range.startDate, definedRange.startDate) && isSameDay(range.endDate, definedRange.endDate);
  },
};

export function createStaticRanges(ranges) {
  return ranges.map((range) => ({ ...staticRangeHandler, ...range }));
}

export const defaultStaticRangesFn = (weekStartsOn) => {
  return createStaticRanges([
    {
      code: "today",
      label: "Today",
      range: () => ({
        startDate: defineds.startOfToday,
        endDate: defineds.endOfToday,
      }),
      // TODO: check if this is being used yet. If not remove all instances of rangeCompare.
      // It was added for a feature that as of this comment hasn't been implemented yet.
      rangeCompare: () => ({
        startDate: defineds.startOfYesterday,
        endDate: defineds.endOfYesterday,
      }),
    },
    {
      code: "yesterday",
      label: "Yesterday",
      range: () => ({
        startDate: defineds.startOfYesterday,
        endDate: defineds.endOfYesterday,
      }),
      rangeCompare: () => ({
        startDate: defineds.startOfYesterdayCompare,
        endDate: defineds.endOfYesterdayCompare,
      }),
    },
    {
      code: "thisweek",
      label: "This Week",
      range: () => ({
        startDate: startOfWeek(new Date(), { weekStartsOn }),
        endDate: endOfWeek(new Date(), { weekStartsOn }),
      }),
      rangeCompare: () => ({
        startDate: startOfWeek(addDays(new Date(), -7), { weekStartsOn }),
        endDate: endOfWeek(addDays(new Date(), -7), { weekStartsOn }),
      }),
    },
    {
      code: "lastweek",
      label: "Last Week",
      range: () => ({
        startDate: startOfWeek(addDays(new Date(), -7), { weekStartsOn }),
        endDate: endOfWeek(addDays(new Date(), -7), { weekStartsOn }),
      }),
      rangeCompare: () => ({
        startDate: endOfWeek(addDays(new Date(), -8), { weekStartsOn }),
        endDate: endOfWeek(addDays(new Date(), -15), { weekStartsOn }),
      }),
    },
    {
      code: "last7days",
      label: "Last 7 Days",
      range: () => ({
        startDate: defineds.startOfLast7Days,
        endDate: defineds.endOfLast7Days,
      }),
      rangeCompare: () => ({
        startDate: defineds.startOfLast7DaysCompare,
        endDate: defineds.endOfLast7DaysCompare,
      }),
    },
    {
      code: "last28days",
      label: "Last 28 Days",
      range: () => ({
        startDate: defineds.last28Days,
        endDate: defineds.startOfToday,
      }),
      rangeCompare: () => ({
        startDate: defineds.startOfLast28DaysCompare,
        endDate: defineds.endOfLast28DaysCompare,
      }),
    },
    {
      code: "last30days",
      label: "Last 30 Days",
      range: () => ({
        startDate: defineds.last30Days,
        endDate: defineds.startOfToday,
      }),
      rangeCompare: () => ({
        startDate: addDays(defineds.last30Days, -30),
        endDate: addDays(defineds.last30Days, -1),
      }),
    },
    {
      code: "yeartodate",
      label: "Year to date",
      range: () => ({
        startDate: defineds.startOfCurrentYear,
        endDate: defineds.startOfToday,
      }),
      rangeCompare: () => ({
        startDate: defineds.startYTDCompare,
        endDate: defineds.endOfYTDCompare,
      }),
    },
    {
      code: "last365days",
      label: "Last 365 days",
      range: () => ({
        startDate: defineds.startOfLast365Days,
        endDate: defineds.startOfToday,
      }),
      rangeCompare: () => ({
        startDate: defineds.startOfLast365DaysCompare,
        endDate: defineds.endOfLast365DaysCompare,
      }),
    },
    {
      code: "alltime",
      label: "All time",
      range: () => ({
        startDate: defineds.startOfTime,
        endDate: defineds.startOfToday,
      }),
      rangeCompare: () => ({
        startDate: defineds.startOfTime,
        endDate: defineds.startOfTime,
      }),
    },
  ]);
};

export const defaultInputRanges = [
  {
    label: "days up to today",
    range(value) {
      return {
        startDate: addDays(defineds.startOfToday, (Math.max(Number(value), 1) - 1) * -1),
        endDate: defineds.endOfToday,
      };
    },
    getCurrentValue(range) {
      if (!isSameDay(range.endDate, defineds.endOfToday)) return "-";
      if (!range.startDate) return "∞";
      return differenceInCalendarDays(defineds.endOfToday, range.startDate) + 1;
    },
  },
  {
    label: "days starting today",
    range(value) {
      const today = new Date();
      return {
        startDate: today,
        endDate: addDays(today, Math.max(Number(value), 1) - 1),
      };
    },
    getCurrentValue(range) {
      if (!isSameDay(range.startDate, defineds.startOfToday)) return "-";
      if (!range.endDate) return "∞";
      return differenceInCalendarDays(range.endDate, defineds.startOfToday) + 1;
    },
  },
];
