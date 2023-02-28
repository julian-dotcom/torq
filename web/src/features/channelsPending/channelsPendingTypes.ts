import { channel } from "features/channels/channelsTypes";

export type ChannelPending = channel & {
  pubKey: string;
  status: string;
};
