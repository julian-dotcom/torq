import { ViewResponse } from "features/viewManagement/types";
import { Peer } from "features/peers/peersTypes";

import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { ColumnMetaData } from "features/table/types";
import { AllPeersColumns, PeersFilterableColumns, PeersSortableColumns } from "./peersColumns.generated";

const defaultColumns: Array<keyof Peer> = ["pubKey", "chain", "network", "actions"];

//changed to not have to rely on the generated file for default columns ordering

//want to add additional columns that are not really part of the model and needs to be there all the time
const extraCols: ColumnMetaData<Peer>[] = [
  {
    heading: "Peer Actions",
    type: "LongTextCell",
    key: "actions",
    valueType: "string",
  },
];
export const PeersColumns: ColumnMetaData<Peer>[] = AllPeersColumns.concat(extraCols);

export const DefaultPeersColumns = defaultColumns.map((col) => PeersColumns.find((x) => x.key === col)) as Array<
  ColumnMetaData<Peer>
>;

export const FilterablePeersColumns = AllPeersColumns.filter((column: ColumnMetaData<Peer>) =>
  PeersFilterableColumns.includes(column.key)
);

export const PeersFilterTemplate: FilterInterface = {
  funcName: "gte",
  category: "number",
  parameter: 0,
  key: "network",
};

export const PeersSortTemplate: { key: keyof Peer; direction: "desc" | "asc" } = {
  key: "pubKey",
  direction: "desc",
};

export const DefaultPeersView: ViewResponse<Peer> = {
  page: "peers",
  dirty: true,
  view: {
    title: "Peers",
    columns: DefaultPeersColumns,
    sortBy: [PeersSortTemplate],
  },
};

export const SortablePeersColumns = AllPeersColumns.filter((c) => PeersSortableColumns.includes(c.key));
