import { Tag } from "pages/tags/tagsTypes";

export type Peer = ExpandedPeer & {
  nodeId: string;
  pubKey: string;
  chain: string;
  network: string;

  tags: Tag[];
};

export type ExpandedPeer = {
  actions: boolean;
};
