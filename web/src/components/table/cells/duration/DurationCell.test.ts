import { formatDurationCell } from "./DurationCell";

describe("formatDurationCell", () => {
  it('formats seconds as "X seconds" when less than 1', () => {
    expect(formatDurationCell(0.5)).toBe("0.500 seconds");
  });

  it('formats seconds as "X and Y" when more than 1', () => {
    expect(formatDurationCell(1234567)).toBe("14 days, 6 hours");
  });

  it('formats seconds as "X and Y" when more than 1 year', () => {
    expect(formatDurationCell(31708800)).toBe("1 year, 2 days");
  });

  it('formats seconds as "X and Y" when more than 1 year', () => {
    expect(formatDurationCell(34387200)).toBe("1 year, 1 month");
  });

  it('formats seconds as "X" when only one unit', () => {
    expect(formatDurationCell(31536000)).toBe("1 year");
    expect(formatDurationCell(2678400)).toBe("1 month"); // 31 days
    expect(formatDurationCell(86400)).toBe("1 day");
    expect(formatDurationCell(3600)).toBe("1 hour");
    expect(formatDurationCell(60)).toBe("1 minute");
    expect(formatDurationCell(1)).toBe("1 second");
  });
});
