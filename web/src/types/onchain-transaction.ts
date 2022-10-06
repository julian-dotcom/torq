export type OnChainTransaction = {
  date: string;
  amount: string;
  destinationAddresses: string[];
  destinationAddressesCount: number;
  label: string;
  lndShortChannelId: string | null;
  lndTransactionTypeLabel: string;
  totalFees: number;
  transactionHash: string;
};
