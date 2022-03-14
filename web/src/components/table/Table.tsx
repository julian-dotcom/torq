import React, {Component} from 'react';
import './table.scss'
import tableRow from "./TableRow";
import NameCell from "./cells/NameCell";
import NumericCell from "./cells/NumericCell";
import BarCell from "./cells/BarCell";

export interface ColumnMetaData {
  primary: string;
  secondary: string;
  type: string;
  width?: string;
  align?: string;
}

interface ColumnMeta {
  primaryHeading: string,
  secondaryHeading: string,
  type?: string,
  key: string,
  width?: number
}

const columns: ColumnMeta[] = [
  {primaryHeading: '', secondaryHeading: 'Name', type: 'NameCell', key: 'group_name'},
  {primaryHeading: '', secondaryHeading: 'Capacity', type:'NumericCell', key: 'capacity'},
  {primaryHeading: '', secondaryHeading: 'Turnover', type:'NumericCell', key: 'turnover'},
  {primaryHeading: 'Forwarded Amount', secondaryHeading: 'Outbound', type:'BarCell', key: 'amount_out'},
  {primaryHeading: 'Forwarded Amount', secondaryHeading: 'Inbound', type:'NumericCell', key: 'amount_in'},
  {primaryHeading: 'Forwarded Amount', secondaryHeading: 'Total', type:'NumericCell', key: 'amount_total'},
  {primaryHeading: 'Forwarding Revenue', secondaryHeading: 'Outbound', type:'NumericCell', key: 'revenue_out'},
  {primaryHeading: 'Forwarding Revenue', secondaryHeading: 'Inbound', type:'NumericCell', key: 'revenue_in'},
  {primaryHeading: 'Forwarding Revenue', secondaryHeading: 'Total', type:'NumericCell', key: 'revenue_total'},
  {primaryHeading: 'Successfull Forwards', secondaryHeading: 'Outbound', type:'NumericCell', key: 'count_out'},
  {primaryHeading: 'Successfull Forwards', secondaryHeading: 'Inbound', type:'NumericCell', key: 'count_in'},
  {primaryHeading: 'Successfull Forwards', secondaryHeading: 'Total', type:'NumericCell', key: 'count_total'},
];

interface RowType {
  group_name: string,
  amount_out: number,
  amount_in: number,
  amount_total: number,
  revenue_out: number,
  revenue_in: number,
  revenue_total: number,
  count_out: number,
  count_in: number,
  count_total: number,
  capacity: number,
  turnover: number,
}

let totalRows: RowType = {
    group_name: "Total",
    amount_out: 1200000,
    amount_in: 1200000,
    amount_total: 1200000,
    revenue_out: 1200000,
    revenue_in: 1200000,
    revenue_total: 1200000,
    count_out: 1200000,
    count_in: 1200000,
    count_total: 1200000,
    capacity: 1200000,
    turnover: 1.42,
}
let currentRows: RowType[] = [
  {
    group_name: "LNBig",
    amount_out: 1200000,
    amount_in: 1200000,
    amount_total: 1200000,
    revenue_out: 1200000,
    revenue_in: 1200000,
    revenue_total: 1200000,
    count_out: 1200000,
    count_in: 1200000,
    count_total: 1200000,
    capacity: 1200000,
    turnover: 1.42,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },
];
let pastRow: RowType[] = [
  {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },
];

function HeaderCell(item: ColumnMeta) {
  return (
    <div className={"header " + item.key} key={item.key}>
      <div className="top">{item.primaryHeading}</div>
      <div className="bottom">{item.secondaryHeading}</div>
    </div>
  )
}

function Table() {
    let key: keyof typeof columns
    let channel: keyof typeof currentRows
    return (
      <div className="table-wrapper">
        <style>
          {".table-content {grid-template-columns: repeat("+Object.keys(columns).length+",  fit-content(200px))}"}
        </style>
        <div className="table-content">
          {columns.map((item) => {
            return  HeaderCell(item)
          })}

          {currentRows.map((currentRow, index) => {
            return columns.map((column) => {
              let key = column.key as keyof RowType
              let past = pastRow[index][key]
              switch (column.type) {
                case 'NameCell':
                  return NameCell((currentRow[key] as string), key, index)
                case 'NumericCell':
                  return NumericCell((currentRow[key] as number), (past as number), key, index)
                case "BarCell":
                  return BarCell((currentRow[key] as number), (totalRows[key] as number), (past as number), key, index)
                default:
                  return NumericCell((currentRow[key] as number), (past as number), key, index)
              }
            })
          })}

        </div>
      </div>
    );
}

export default Table;
