import React from 'react';
import './table.scss'

export interface ColumnMetaData {
  primary: string;
  secondary: string;
  type: string;
  width?: string;
  align?: string;
}

const headers = {
  "group_name": {primary: '', secondary: 'Name', type: ''},
  "amount_out": {primary: 'Amount', secondary: 'Outbound', type: ''},
  "amount_in": {primary: 'Amount', secondary: 'Inbound', type: ''},
  "amount_total": {primary: 'Amount', secondary: 'Total', type: ''},
  "revenue_out": {primary: 'Revenue', secondary: 'Outbound', type: ''},
  "revenue_in": {primary: 'Revenue', secondary: 'Inbound', type: ''},
  "revenue_total": {primary: 'Revenue', secondary: 'Total', type: ''},
  "count_out": {primary: 'Count', secondary: 'Outbound', type: ''},
  "count_in": {primary: 'Count', secondary: 'Inbound', type: ''},
  "count_total": {primary: 'Count', secondary: 'Total', type: ''},
  "capacity": {primary: '', secondary: 'Capacity', type: ''},
  "turnover": {primary: '', secondary: 'Turnover', type: ''},
}

let currentRows = [
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

let previousRows = [
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

function HeaderCell(key: string, primary :string, secondary:string) {
  return (
    <div className={"header"}>
      <div className="top">{primary}</div>
      <div className="bottom">{secondary}</div>
    </div>
  )
}

function Table() {
    let key: keyof typeof headers
    let channel: keyof typeof currentRows
    return (
      <div className="table-wrapper">
        <style>
          {".table-content {grid-template-columns: repeat("+Object.keys(headers).length+", minmax(100px,  1fr))}"}
        </style>
        <div className="table-content">
          {Object.entries(headers).map(([key, item]) => {
            console.log(key)
            return  HeaderCell(key, item.primary,item.secondary)

          })}

          {currentRows.map((row, index) => {
            return Object.entries(row).map(([key, cell]) => {
              return (
                <div className={"cell " + key} key={key}>
                  {cell}
                </div>
              )
            })
          })}

        </div>
      </div>
    );
}

export default Table;
