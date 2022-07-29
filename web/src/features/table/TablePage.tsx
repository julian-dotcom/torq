import styles from "./table-page.module.scss";
import TableControls from "./controls/TableControls";
import Table from "./tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice";
import TablePageTemplate from "../templates/TablePageTemplate";
import { Link } from "react-router-dom";

function TablePage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery();

  const breadcrumbs = ["Analyse", <Link to={"forwards"}>Forwards</Link>];
  return (
    <TablePageTemplate title={"Forwards"} breadcrumbs={breadcrumbs}>
      <div className={styles.tablePageWrapper}>
        <Table />
      </div>
    </TablePageTemplate>
  );
}

export default TablePage;
