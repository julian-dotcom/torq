import { channel } from "features/channels/channelsTypes";

export type ChannelClosed = channel & {
  pubKey: string;
  status: string;
  closingNodeName: string;
};
