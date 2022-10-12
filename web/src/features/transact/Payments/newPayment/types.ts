export type NewPaymentRequest = {
  invoice: string;
  timeOutSecs: number;
  dest: string;
  amtMSat: number;
  feeLimitMsat: number;
  allowSelfPayment: boolean;
};

export enum PaymentType {
  Unknown,
  Keysend,
  P2PKH,
  P2SH,
  P2WKH,
  P2TR,
  LightningMainnet,
  LightningTestnet,
  LightningSimnet,
  LightningRegtest,
}
