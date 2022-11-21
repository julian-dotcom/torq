import { Pagination } from "types/api";

export type PaymentsResponse = {
  data: Payment[];
  pagination: Pagination;
};

export type Payment = {
  paymentIndex: number;
  date: string;
  destinationPubKey: string;
  status: string;
  value: number;
  fee: number;
  ppm: number;
  failureReason: string;
  paymentHash: number;
  paymentPreimage: number;
  isRebalance: boolean;
  isMpp: boolean;
  countFailedAttempts: number;
  countSuccessfulAttempts: number;
  secondsInFlight: number;
};
