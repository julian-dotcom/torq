import styles from "../table/table-page.module.scss";
import TableControls from "../table/controls/TableControls";
import Table from "../table/tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice";
import TablePageTemplate from "../templates/TablePageTemplate";
import { Link } from "react-router-dom";

function PaymentsPage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery();

  const breadcrumbs = ["Transact", <Link to={"/transact/payments"}>Payments</Link>];
  return (
    <TablePageTemplate title={"Transactions"} breadcrumbs={breadcrumbs} sidebarExpanded={false}>
      <div className={styles.tablePageWrapper}>
        <Table />
      </div>
    </TablePageTemplate>
  );
}

export default PaymentsPage;
