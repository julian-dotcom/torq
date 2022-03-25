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
  differenceInCalendarDays,
} from 'date-fns';

export const defineds = {
  startOfWeek: startOfWeek(new Date()),
  endOfWeek: endOfWeek(new Date()),
  startOfWeekCompare: startOfWeek(addDays(new Date(), -7)),
  endOfWeekCompare: endOfWeek(addDays(new Date(), -7)),
  startOfLastWeek: startOfWeek(addDays(new Date(), -7)),
  endOfLastWeek: endOfWeek(addDays(new Date(), -7)),
  startOfLastWeekCompare: endOfWeek(addDays(new Date(), -8)),
  endOfLastWeekCompare: endOfWeek(addDays(new Date(), -15)),
  startOfToday: startOfDay(new Date()),
  endOfToday: endOfDay(new Date()),
  startOfYesterday: startOfDay(addDays(new Date(), -1)),
  endOfYesterday: endOfDay(addDays(new Date(), -1)),
  startOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),
  endOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),
  startOfMonth: startOfMonth(new Date()),
  endOfMonth: endOfMonth(new Date()),
  startOfLastMonth: startOfMonth(addMonths(new Date(), -1)),
  endOfLastMonth: endOfMonth(addMonths(new Date(), -1)),
  last28Days: startOfDay(addDays(new Date(), -28)),
  startOfLast28DaysCompare: startOfDay(addDays(new Date(), -29)),
  endOfLast28DaysCompare: startOfDay(addDays(new Date(), -57)),
  last30Days: startOfDay(addDays(new Date(), -30)),
  startOfLast7Days: startOfDay(new Date()),
  endofLast7Days: startOfDay(addDays(new Date(), -7)),
  startOfLast7DaysCompare: startOfDay(addDays(new Date(), -8)),
  endOfLast7DaysCompare: startOfDay(addDays(new Date(), -15))
};


export const getCompareRanges = (startDate, endDate) => {
  const daysDifference = differenceInCalendarDays(endDate, startDate);
  let compareRange = [
    addDays(startOfDay(startDate), -1),
    addDays(startOfDay(startDate), -(daysDifference + 1)),
  ];

  return compareRange
}

const staticRangeHandler = {
  range: {},
  isSelected(range) {
    const definedRange = this.range();
    return (
      isSameDay(range.startDate, definedRange.startDate) &&
      isSameDay(range.endDate, definedRange.endDate)
    );
  },
};

export function createStaticRanges(ranges) {
  return ranges.map((range) => ({ ...staticRangeHandler, ...range }));
}

export const defaultStaticRanges = createStaticRanges([
  {
    label: 'Today',
    range: () => ({
      startDate: defineds.startOfToday,
      endDate: defineds.endOfToday,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfYesterday,
      endDate: defineds.endOfYesterday,
    })
  },
  {
    label: 'Yesterday',
    range: () => ({
      startDate: defineds.startOfYesterday,
      endDate: defineds.endOfYesterday,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfYesterdayCompare,
      endDate: defineds.endOfYesterdayCompare,
    })
  },
  {
    label: 'This Week',
    range: () => ({
      startDate: defineds.startOfWeek,
      endDate: defineds.endOfWeek,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfWeekCompare,
      endDate: defineds.endOfLastWeekCompare,
    }),
  },
  {
    label: 'Last Week',
    range: () => ({
      startDate: defineds.startOfLastWeek,
      endDate: defineds.endOfLastWeek,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLastWeekCompare,
      endDate: defineds.endOfLastWeekCompare,
    })
  },
  {
    label: 'Last 7 Days',
    range: () => ({
      startDate: defineds.startOfToday,
      endDate: defineds.endofLast7Days,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast7DaysCompare,
      endDate: defineds.endOfLast7DaysCompare,
    }),
  },
  {
    label: 'Last 28 Days',
    range: () => ({
      startDate: defineds.startOfToday,
      endDate: defineds.last28Days,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast28DaysCompare,
      endDate: defineds.endOfLast28DaysCompare,
    })
  },
  {
    label: 'Last 30 Days',
    range: () => ({
      startDate: defineds.startOfToday,
      endDate: defineds.last30Days,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast28DaysCompare,
      endDate: defineds.endOfLast28DaysCompare,
    })
  },
]);

export const defaultInputRanges = [
  {
    label: 'days up to today',
    range(value) {
      return {
        startDate: addDays(defineds.startOfToday, (Math.max(Number(value), 1) - 1) * -1),
        endDate: defineds.endOfToday,
      };
    },
    getCurrentValue(range) {
      if (!isSameDay(range.endDate, defineds.endOfToday)) return '-';
      if (!range.startDate) return '∞';
      return differenceInCalendarDays(defineds.endOfToday, range.startDate) + 1;
    },
  },
  {
    label: 'days starting today',
    range(value) {
      const today = new Date();
      return {
        startDate: today,
        endDate: addDays(today, Math.max(Number(value), 1) - 1),
      };
    },
    getCurrentValue(range) {
      if (!isSameDay(range.startDate, defineds.startOfToday)) return '-';
      if (!range.endDate) return '∞';
      return differenceInCalendarDays(range.endDate, defineds.startOfToday) + 1;
    },
  },
];
