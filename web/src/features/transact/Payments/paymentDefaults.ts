import { ColumnMetaData } from "features/table/types";
import { Payment } from "./types";
import { ViewInterface } from "../../viewManagement/types";

export const AllPaymentsColumns: Array<ColumnMetaData<Payment>> = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "status", heading: "Status", type: "TextCell", valueType: "array" },
  { key: "value", heading: "Value", type: "NumericCell", valueType: "number" },
  { key: "fee", heading: "Fee", type: "NumericCell", valueType: "number" },
  { key: "ppm", heading: "PPM", type: "NumericCell", valueType: "number" },
  { key: "isRebalance", heading: "Rebalance", type: "BooleanCell", valueType: "boolean" },
  { key: "secondsInFlight", heading: "Seconds In Flight", type: "DurationCell", valueType: "duration" },
  { key: "failureReason", heading: "Failure Reason", type: "TextCell", valueType: "array" },
  { key: "isMpp", heading: "MPP", type: "BooleanCell", valueType: "boolean" },
  { key: "countFailedAttempts", heading: "Failed Attempts", type: "NumericCell", valueType: "number" },
  { key: "countSuccessfulAttempts", heading: "Successful Attempts", type: "NumericCell", valueType: "number" },
  { key: "destinationPubKey", heading: "Destination", type: "TextCell", valueType: "string" },
  { key: "paymentHash", heading: "Payment Hash", type: "TextCell", valueType: "string" },
  { key: "paymentPreimage", heading: "Payment Preimage", type: "TextCell", valueType: "string" },
];

export const defaultColumns: Array<keyof Payment> = [
  "date",
  "status",
  "value",
  "fee",
  "ppm",
  "isRebalance",
  "secondsInFlight",
  "failureReason",
  "countFailedAttempts",
];

export const ActivePaymentsColumns: Array<ColumnMetaData<Payment>> = AllPaymentsColumns.filter((item) => {
  return defaultColumns.includes(item.key);
});

export const DefaultPaymentView: ViewInterface<Payment> = {
  title: "Untitled View",
  saved: true,
  columns: ActivePaymentsColumns,
  page: "payments",
  sortBy: [],
  view_order: 0,
};
