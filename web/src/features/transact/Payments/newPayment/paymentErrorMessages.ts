export const ProcessingPaymentErrors = new Map<string, string>([
  ["ALREADY_PAID", "Invoice is already paid"],
  ["INVALID_HASH", "Invalid invoice (invalid hash)"],
  ["INVALID_PAYMENT_REQUEST", "Invalid invoice"],
  ["CHECKSUM_FAILED", "Invalid invoice (checksum failed)"],
  ["UNKNOWN_ERROR", "Unknown error occurred"],
]);
