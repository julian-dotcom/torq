export type NewInvoiceRequest = {
  nodeId: number;
  memo?: string;
  rPreImage?: string;
  valueMsat?: number;
  expiry?: number;
  fallBackAddress?: string;
  private?: boolean;
  isAmp?: boolean;
};

export type NewInvoiceResponse = {
  nodeId: number;
  paymentRequest: string;
  addIndex: number;
  paymentAddress: string;
};

export type NewInvoiceError = { id: string; type: "Error"; error: string };
