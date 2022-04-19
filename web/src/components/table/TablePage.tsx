import './table-page.scss'
import { useEffect } from 'react';
import { format } from "date-fns";

import TableControls from "./controls/TableControls";
import Table from "./tableContent/Table";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import { fetchChannelsAsync, } from "./tableSlice";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";

function TablePage() {
  const dispatch = useAppDispatch();
  const currentPeriod = useAppSelector(selectTimeInterval);

  useEffect(() => {
    const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
    const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
    dispatch(fetchChannelsAsync({ from: from, to: to }));
  }, [currentPeriod])

  return (
    <div className="table-page-wrapper">
      <div className="table-controls-wrapper">
        <TableControls />
      </div>
      <Table />
    </div>
  );
}

export default TablePage;
