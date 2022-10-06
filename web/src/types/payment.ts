export type Payment = {
  paymentIndex: number;
  date: string;
  destinationPubKey: string;
  status: string;
  value: number;
  fee: number;
  ppm: number;
  failureReason: string;
  paymentHash: string;
  paymentPreimage: string;
  paymentRequest: string;
  isRebalance: boolean;
  isMpp: boolean;
  countSuccessfulAttempts: number;
  countFailedAttempts: number;
  secondsInFlight: number;
};
