import styles from "./table-page.module.scss";
import TableControls from "./controls/TableControls";
import Table from "./tableContent/Table";
import { useGetTableViewsQuery } from "apiSlice";
import TablePageTemplate from "../templates/TablePageTemplate";
import { Link } from "react-router-dom";
import SidebarSection from "../sidebar/SidebarSection";
import { Filter20Regular as FilterIcon, ArrowSortDownLines20Regular as SortIcon } from "@fluentui/react-icons";

function TablePage() {
  // initial getting of the table views from the database
  useGetTableViewsQuery();

  const breadcrumbs = ["Analyse", <Link to={"forwards"}>Forwards</Link>];
  return (
    <TablePageTemplate
      title={"Forwards"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={true}
      sidebarChildren={
        <div>
          <SidebarSection title={"Filter"} icon={FilterIcon} collapsed={false}>
            {"Something"}
          </SidebarSection>
          <SidebarSection title={"Sort"} icon={SortIcon} collapsed={true}>
            {"Something"}
          </SidebarSection>
        </div>
      }
    >
      <div className={styles.tablePageWrapper}>
        <Table />
      </div>
    </TablePageTemplate>
  );
}

export default TablePage;
