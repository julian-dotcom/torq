import { ChannelStatus } from "constants/backend";
import { channel } from "features/channels/channelsTypes";

export type ChannelClosed = channel & {
  pubKey: string;
  status: ChannelStatus;
  closingNodeName: string;
};
