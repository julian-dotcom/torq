import React, {useEffect} from 'react';
import TableControls from "../components/table/TableControls";
import Table from "../components/table/Table";
import './table-page.scss'
import {useAppDispatch, useAppSelector} from "../store/hooks";
import {fetchChannelsAsync, fetchTableViewsAsync} from "../components/table/tableSlice";
import {selectTimeInterval} from "../components/timeIntervalSelect/timeIntervalSlice";
import {format} from "date-fns";

function TablePage() {
  const dispatch = useAppDispatch();
  const currentPeriod = useAppSelector(selectTimeInterval);

  useEffect(() =>{
    dispatch(fetchTableViewsAsync());

    const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
    const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
    dispatch(fetchChannelsAsync({ from: from, to: to }));
  })



  return (
    <div className="table-page-wrapper">
      <div className="table-controls-wrapper">
        <TableControls/>
      </div>
      <Table/>
    </div>
  );
}

export default TablePage;
