import { ColumnMetaData } from "features/table/types";
import { Payment } from "features/transact/Payments/types";
import { ViewResponse } from "features/viewManagement/types";
import { FilterCategoryType, FilterInterface } from "features/sidebar/sections/filter/filter";
import { AllPaymentsColumns, PaymentsSortableColumns, PaymentsFilterableColumns } from "features/transact/Payments/paymentsColumns";

const defaultColumns: Array<keyof Payment> = [
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

export const SortablePaymentsColumns = AllPaymentsColumns.filter((column: ColumnMetaData<Payment>) =>
  PaymentsSortableColumns.includes(column.key)
);

export const FilterablePaymentsColumns = AllPaymentsColumns.filter((column: ColumnMetaData<Payment>) =>
  PaymentsFilterableColumns.includes(column.key)
);

export const PaymentsFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number" as FilterCategoryType,
  parameter: 0,
  key: "value",
};

export const StatusTypeLabels = new Map<string, string>([
  ["SUCCEEDED", "Succeeded"],
  ["FAILED", "Failed"],
  ["IN_FLIGHT", "In Flight"],
]);

export const FailureReasonLabels = new Map<string, string>([
  ["FAILURE_REASON_NONE", ""],
  ["FAILURE_REASON_TIMEOUT", "Timeout"],
  ["FAILURE_REASON_NO_ROUTE", "No Route"],
  ["FAILURE_REASON_ERROR", "Error"],
  ["FAILURE_REASON_INCORRECT_PAYMENT_DETAILS", "Incorrect Payment Details"],
  ["FAILURE_REASON_INCORRECT_PAYMENT_AMOUNT", "Incorrect Payment Amount"],
  ["FAILURE_REASON_PAYMENT_HASH_MISMATCH", "Payment Hash Mismatch"],
  ["FAILURE_REASON_INCORRECT_PAYMENT_REQUEST", "Incorrect Payment Request"],
  ["FAILURE_REASON_UNKNOWN", "Unknown"],
]);

export const PaymentsSortTemplate: { key: keyof Payment; direction: "desc" | "asc" } = {
  key: "date",
  direction: "desc",
};

export const ActivePaymentsColumns: Array<ColumnMetaData<Payment>> = AllPaymentsColumns.filter((item) => {
  return defaultColumns.includes(item.key);
});

export const DefaultPaymentView: ViewResponse<Payment> = {
  page: "payments",
  dirty: true,
  view: {
    title: "Draft View",
    columns: ActivePaymentsColumns,
    sortBy: [PaymentsSortTemplate],
  },
};
