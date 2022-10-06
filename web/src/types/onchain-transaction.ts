export type OnChainTransaction = {
  date: string;
  amount: string;
  destinationAddresses: string[];
  destinationAddressesCount: number;
  label: string;
  lndShortChannelId: string | null;
  lndTxTypeLabel: string;
  totalFees: number;
  txHash: string;
};
