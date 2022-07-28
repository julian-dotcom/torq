import styles from "../table/table-page.module.scss";
import TableControls from "../table/controls/TableControls";
import Table from "../table/tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice";
import TablePageTemplate from "../templates/TablePageTemplate";

function PaymentsPage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery();

  return (
    <TablePageTemplate title={"Transactions"}>
      <div className={styles.tablePageWrapper}>
        <Table />
      </div>
    </TablePageTemplate>
  );
}

export default PaymentsPage;
