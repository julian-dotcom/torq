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
  differenceInCalendarDays, addYears,
} from 'date-fns';
// import locale from 'date-fns/locale/en-US'
import locale from 'date-fns/locale/nb'

export const defineds = {
  startOfToday: startOfDay(new Date()),
  endOfToday: endOfDay(new Date()),

  startOfYesterday: startOfDay(addDays(new Date(), -1)),
  endOfYesterday: endOfDay(addDays(new Date(), -1)),
  startOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),
  endOfYesterdayCompare: startOfDay(addDays(new Date(), -2)),


  startOfLast7Days: startOfDay(addDays(new Date(), -7)),
  endOfLast7Days: startOfDay(new Date()),
  startOfLast7DaysCompare: startOfDay(addDays(new Date(), -8)),
  endOfLast7DaysCompare: startOfDay(addDays(new Date(), -15)),

  startOfWeek: startOfWeek(new Date(), {locale}),
  endOfWeek: endOfWeek(new Date(), {locale}),
  startOfWeekCompare: startOfWeek(addDays(new Date(), -7), {locale}),
  endOfWeekCompare: endOfWeek(addDays(new Date(), -7), {locale}),

  startOfLastWeek: startOfWeek(addDays(new Date(), -7), {locale}),
  endOfLastWeek: endOfWeek(addDays(new Date(), -7), {locale}),
  startOfLastWeekCompare: endOfWeek(addDays(new Date(), -8), {locale}),
  endOfLastWeekCompare: endOfWeek(addDays(new Date(), -15), {locale}),


  startOfMonth: startOfMonth(new Date()),
  endOfMonth: endOfMonth(new Date()),
  startOfLastMonth: startOfMonth(addMonths(new Date(), -1)),
  endOfLastMonth: endOfMonth(addMonths(new Date(), -1)),

  last28Days: startOfDay(addDays(new Date(), -28)),
  startOfLast28DaysCompare: startOfDay(addDays(new Date(), -57)),
  endOfLast28DaysCompare: startOfDay(addDays(new Date(), -29)),

  last30Days: startOfDay(addDays(new Date(), -30)),

  startOfLast365Days: startOfDay(addYears(new Date(), -1)),

  startOfCurrentYear: startOfYear(new Date()),
  startYTDCompare: addYears(startOfYear(new Date()), -1),
  endOfYTDCompare: addYears(new Date(), -1),

  startOfLast365DaysCompare: startOfDay(addYears(new Date(), -2)),
  endOfLast365DaysCompare: startOfDay(addYears(addDays(new Date(), -1), -1)),

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
      startDate: defineds.startOfLast7Days,
      endDate: defineds.endOfLast7Days,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast7DaysCompare,
      endDate: defineds.endOfLast7DaysCompare,
    }),
  },
  {
    label: 'Last 28 Days',
    range: () => ({
      startDate: defineds.last28Days,
      endDate: defineds.startOfToday,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast28DaysCompare,
      endDate: defineds.endOfLast28DaysCompare,
    })
  },
  {
    label: 'Last 30 Days',
    range: () => ({
      startDate: defineds.last30Days,
      endDate: defineds.startOfToday,
    }),
    rangeCompare: () => ({
      startDate: addDays(defineds.last30Days, -30),
      endDate: addDays(defineds.last30Days, -1),
    })
  },
  {
    label: 'Year to date',
    range: () => ({
      startDate: defineds.startOfCurrentYear,
      endDate: defineds.startOfToday,
    }),
    rangeCompare: () => ({
      startDate: defineds.startYTDCompare,
      endDate: defineds.endOfYTDCompare,
    })
  },
  {
    label: 'Last 365 days',
    range: () => ({
      startDate: defineds.startOfLast365Days,
      endDate: defineds.startOfToday,
    }),
    rangeCompare: () => ({
      startDate: defineds.startOfLast365DaysCompare,
      endDate: defineds.endOfLast365DaysCompare,
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
