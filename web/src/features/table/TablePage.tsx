import styles from './table-page.module.scss'
import { useEffect } from 'react';
import TableControls from "./controls/TableControls";
import Table from "./tableContent/Table";
import { useAppDispatch } from "../../store/hooks";
import { fetchTableViewsAsync } from "./tableSlice";

function TablePage() {

  const dispatch = useAppDispatch();

  useEffect(() =>{
    dispatch(fetchTableViewsAsync());
  })

  return (
    <div className={styles.tablePageWrapper}>
      <div className="table-controls-wrapper">
        <TableControls />
      </div>
      <Table />
    </div>
  );
}

export default TablePage;
