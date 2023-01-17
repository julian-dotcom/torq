import { Tag } from "pages/tags/tagsTypes";

export type Htlc = {
  eventTime: Date;
  eventType: string;
  eventOrigin: string;

  incomingChannelId: number;
  incomingShortChannelId: string;
  incomingLndShortChannelId: string;
  incomingChannelTags: Array<Tag>;
  incomingAlias: string;
  incomingPublicKey: string;
  incomingNodeId: number;
  incomingChannelStatus: number;
  incomingChannelCapacity: number;
  incomingTimeLock: number;
  incomingAmountMsat: number;

  outgoingChannelId: number;
  outgoingShortChannelId: string;
  outgoingLndShortChannelId: string;
  outgoingChannelTags: Array<Tag>;
  outgoingAlias: string;
  outgoingPublicKey: string;
  outgoingNodeId: number;
  outgoingChannelStatus: number;
  outgoingChannelCapacity: number;
  outgoingTimeLock: number;
  outgoingAmountMsat: number;

  boltFailureCode: string;
  boltFailureString: string;
  lndFailureDetail: string;

  nodeId: number;
};
