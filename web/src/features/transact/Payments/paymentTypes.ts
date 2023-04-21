type MppRecord = {
  paymentAddr: string;
  totalAmtMsat: number;
};

type hops = {
  chanId: string;
  expiry: number;
  amtToForwardMsat: number;
  pubKey: string;
  mppRecord: MppRecord;
  // TODO: Imolement AMP record here when needed
};

type route = {
  totalTimeLock: number;
  hops: Array<hops>;
  totalAmtMsat: number;
};

type failureDetails = {
  reason: string;
  failureSourceIndex: number;
  height: number;
};

type attempt = {
  attemptId: number;
  status: string;
  route: route;
  attemptTimeNs: Date;
  resolveTimeNs: Date;
  preimage: string;
  failure: failureDetails;
};

export type InvoiceStatusType = "SUCCEEDED" | "FAILED" | "IN_FLIGHT";

export type NewPaymentResponse = {
  type: string;
  status: InvoiceStatusType;
  hash: string;
  preimage: string;
  paymentRequest: string;
  amountMsat: number;
  feeLimitMsat: number;
  feePaidMsat: number;
  creationDate: Date;
  failureReason: string;
  attempt: attempt;
};
