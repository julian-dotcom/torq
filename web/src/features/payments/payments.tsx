import styles from "../table/table-page.module.scss";
import TableControls from "../table/controls/TableControls";
import Table from "../table/tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice";

function PaymentsPage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery();

  return (
    <div className={styles.tablePageWrapper}>
      <div className="table-controls-wrapper">
        <TableControls />
      </div>
      <Table />
    </div>
  );
}

export default PaymentsPage;
