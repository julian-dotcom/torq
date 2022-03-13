import React from 'react';
import './table_controls.scss'
import DefaultButton from "../buttons/Button";
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDown20Regular as SortIcon,
  Filter20Filled as FilterIcon,
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";


function TableControls() {
    return (
      <div className="table-controls">

        <div className="left-container">
          <div className="title">
            <h1>Top revenue today</h1>
          </div>
          <TimeIntervalSelect/>
        </div>

        <div className="right-container">
          <DefaultButton icon={<ColumnsIcon/>} text={"Columns"}/>
          <DefaultButton icon={<SortIcon/>} text={"Sort"}/>
          <DefaultButton icon={<FilterIcon/>} text={"Filter"}/>
        </div>

      </div>
    );
}

export default TableControls;
