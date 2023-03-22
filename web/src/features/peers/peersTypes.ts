import { Tag } from "pages/tags/tagsTypes";

export type Peer = {
  nodeId: string;
  pubKey: string;
  chain: string;
  network: string;

  tags: Tag[];
};
