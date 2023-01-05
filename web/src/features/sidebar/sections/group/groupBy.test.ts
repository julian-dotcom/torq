import { useGroupBy } from "./groupBy";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const testData: Array<any> = [
  {
    alias: "Some Node",
    channelId: 32,
    channel_point: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e288:0",
    pub_key: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2052:0",
    chan_id: "769235927112613888",
    color: "#68f442",
    open: 1,
    capacity: 10000000,
    count_total: 20,
    htlc_fail_all_in: 10,
    htlc_fail_all_out: 150,
  },
  {
    alias: "Another Node",
    channelId: 40,
    channel_point: "f1c17e33b03bb3722eee187d5cceaaeab7b1e3e72d6efcbebf263747122a770f:0",
    pub_key: "033f405aae705d96d4338efb236645a61c9b0a2303e3185211ed3b02c0803a4a2a",
    shortChannelId: "707781:900:1",
    chan_id: "778213439477907457",
    color: "#68f4a2",
    open: 1,
    capacity: 2000000,
    count_total: 10,
    htlc_fail_all_in: 5,
    htlc_fail_all_out: 50,
  },
  {
    alias: "Some Node",
    channelId: 33,
    channel_point: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e289:0",
    pub_key: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2053:0",
    chan_id: "769235927112613889",
    color: "#68f442",
    open: 0,
    capacity: 5000000,
    count_total: 10,
    htlc_fail_all_in: 10,
    htlc_fail_all_out: 100,
  },
  {
    alias: "Some Node",
    channelId: 34,
    channel_point: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e290:0",
    pub_key: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2054:0",
    chan_id: "769235927112613890",
    color: "#68f442",
    open: 0,
    capacity: 5000000,
    count_total: 0,
    htlc_fail_all_in: 0,
    htlc_fail_all_out: 0,
  },
];

test("Unknown by param returns exactly what was input", () => {
  const result = useGroupBy(testData, "");

  expect(result).toStrictEqual(testData);
});

test("grouping by channels returns exactly what was input", () => {
  const result = useGroupBy(testData, "channel");

  expect(result).toStrictEqual(testData);
});

test("grouping by peers returns correctly grouped channels", () => {
  const result = useGroupBy(testData, "peers");

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const expected: Array<any> = [
    {
      alias: "Some Node",
      channelId: [32, 33, 34],
      channel_point: [
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e288:0",
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e289:0",
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e290:0",
      ],
      pub_key: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
      shortChannelId: ["699616:2052:0", "699616:2053:0", "699616:2054:0"],
      chan_id: ["769235927112613888", "769235927112613889", "769235927112613890"],
      color: "#68f442",
      open: 1,
      capacity: 20000000,
      count_total: 30,
      htlc_fail_all_in: 20,
      htlc_fail_all_out: 250,
    },
    {
      alias: "Another Node",
      channelId: 40,
      channel_point: "f1c17e33b03bb3722eee187d5cceaaeab7b1e3e72d6efcbebf263747122a770f:0",
      pub_key: "033f405aae705d96d4338efb236645a61c9b0a2303e3185211ed3b02c0803a4a2a",
      shortChannelId: "707781:900:1",
      chan_id: "778213439477907457",
      color: "#68f4a2",
      open: 1,
      capacity: 2000000,
      count_total: 10,
      htlc_fail_all_in: 5,
      htlc_fail_all_out: 50,
    },
  ];

  expect(result).toStrictEqual(expected);
});
