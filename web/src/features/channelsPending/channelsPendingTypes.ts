import { channel } from "features/channels/channelsTypes";

export type ChannelPending = Omit<channel, "tag"> & {
  pubKey: string;
  status: string;
};
