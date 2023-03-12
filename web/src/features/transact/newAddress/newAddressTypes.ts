export enum AddressType {
  P2WPKH = 1,
  P2WKH = 2,
  P2TR = 4,
}

export type NewAddressRequest = {
  nodeId: number;
  addressType: AddressType;
};

export type NewAddressResponse = {
  type: AddressType;
  address: string;
};

export type NewAddressError = { id: string; type: "Error"; error: string };
