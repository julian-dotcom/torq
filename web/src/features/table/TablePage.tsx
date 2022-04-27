import styles from './table-page.module.scss'
import TableControls from "./controls/TableControls";
import Table from "./tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice"

function TablePage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery()

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
