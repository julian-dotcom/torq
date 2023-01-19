export type SignMessageRequest = {
  nodeId: number;
  message: string;
};

export type SignMessageResponse = {
  signature: string;
};

export type VerifyMessageRequest = {
  nodeId: number;
  message: string;
  signature: string;
};

export type VerifyMessageResponse = {
  valid: boolean;
  pubKey: string;
};
