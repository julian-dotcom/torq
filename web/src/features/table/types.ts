export type ColumnMetaData<T extends {}> = {
  heading: string;
  key: keyof T;
  type?: string;
  width?: number;
  locked?: boolean;
  valueType: string;
  total?: number;
  max?: number;
  percent?: boolean;
};

export type TableProps<T extends {}> = {
  activeColumns: Array<ColumnMetaData<T>>;
  data: Array<T>;
  isLoading: boolean;
  showTotals?: boolean;
  totalRow?: Array<T>;
  cellRenderer: CellRendererFunction<T>;
  selectable?: boolean;
  selectedRowIds?: Array<number>;
};

export type CellRendererFunction<T extends {}> = (
  row: T,
  index: number,
  columnMeta: ColumnMetaData<T>,
  columnIndex: number
) => JSX.Element;

export type RowProp<T extends {}> = {
  row: T;
  rowIndex: number;
  columns: Array<ColumnMetaData<T>>;
  selectable?: boolean;
  selected: boolean;
  cellRenderer: CellRendererFunction<T>;
  isTotalsRow?: boolean;
};
