import './table-page.scss'
import { useEffect } from 'react';
import { format } from "date-fns";

import TableControls from "../components/table/TableControls";
import Table from "../components/table/Table";
import { useAppDispatch, useAppSelector } from "../store/hooks";
import { fetchChannelsAsync, } from "../components/table/tableSlice";
import { selectTimeInterval } from "../components/timeIntervalSelect/timeIntervalSlice";

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
