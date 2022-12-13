import { Pagination } from "types/api";

export type OnChainResponse = {
  data: OnChainTx[];
  pagination: Pagination;
};

export type OnChainTx = {
  date: string;
  amount: number;
  destAddresses: string[];
  destAddressesCount: string;
  label: string;
  lndShortChanId: string;
  lndTxTypeLabel: string;
  totalFees: number;
  txHash: number;
};
