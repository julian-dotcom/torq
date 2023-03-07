import { channel } from "features/channels/channelsTypes";

export type ChannelClosed = Omit<channel, "tag"> & {
  pubKey: string;
  status: string;
  closingNodeName: string;
};
