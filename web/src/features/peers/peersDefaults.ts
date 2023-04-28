import { ViewResponse } from "features/viewManagement/types";
import { Peer } from "features/peers/peersTypes";

import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import { AllPeersColumns, PeersFilterableColumns, PeersSortableColumns } from "./peersColumns.generated";
import { OrderBy } from "../sidebar/sections/sort/SortSection";

const defaultColumns: Array<keyof Peer> = ["peerAlias", "connectionStatus", "nodeName", "pubKey", "tags"];

export const DefaultPeersColumns = defaultColumns.map((col) => AllPeersColumns.find((x) => x.key === col)) as Array<
  ColumnMetaData<Peer>
>;

export const FilterablePeersColumns = AllPeersColumns.filter((column: ColumnMetaData<Peer>) =>
  PeersFilterableColumns.includes(column.key)
);

export const PeersFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: "0",
  key: "value",
};

export const PeersSortTemplate: Array<OrderBy> = [
  {
    key: "peerAlias",
    direction: "desc",
  },
];

export const DefaultPeersView: ViewResponse<Peer> = {
  page: "peers",
  dirty: true,
  view: {
    title: "Peers",
    columns: DefaultPeersColumns,
    sortBy: PeersSortTemplate,
  },
};

export const SortablePeersColumns = AllPeersColumns.filter((c) => PeersSortableColumns.includes(c.key));
