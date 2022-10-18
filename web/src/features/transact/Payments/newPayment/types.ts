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

export const PaymentTypeLabel = {
  [PaymentType.Unknown]: "Unknown ",
  [PaymentType.P2PKH]: "Legacy Bitcoin ", // Legacy address
  [PaymentType.P2SH]: "Pay-to-Script-Hash ", // P2SH address
  [PaymentType.P2WKH]: "Segwit ", // Segwit address
  [PaymentType.P2TR]: "Taproot Address", // Taproot address
  [PaymentType.LightningMainnet]: "Mainnet Invoice",
  [PaymentType.LightningTestnet]: "Testnet Invoice",
  [PaymentType.LightningSimnet]: "Simnet Invoice",
  [PaymentType.LightningRegtest]: "Regtest Invoice",
  [PaymentType.Keysend]: "Keysend",
};
