import Table from "features/table/Table";
// import { ColumnMetaData } from "features/table/types";
import { useGetInvoicesQuery } from "apiSlice";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Options20Regular as OptionsIcon,
  Check20Regular as InvoiceIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Pagination from "components/table/pagination/Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";
// import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
// import FilterSection from "features/sidebar/sections/filter/FilterSection";
// import { Clause, FilterInterface } from "features/sidebar/sections/filter/filter";
// import { useAppDispatch, useAppSelector } from "store/hooks";
// import { FilterCategoryType } from "features/sidebar/sections/filter/filter";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import { SectionContainer } from "features/section/SectionContainer";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_INVOICE } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import { AllViewsResponse } from "features/viewManagement/types";
import { InvoicesResponse } from "./invoiceTypes";
import {
  AllInvoicesColumns,
  DefaultFilter,
  DefaultSortValue,
  InvoiceViewTemplate,
  SortableInvoiceColumns,
} from "./invoiceDefaults";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";

import SortSection from "../../sidebar/sections/sort/SortSection";
import { useView } from "../../viewManagement/useView";
import FilterSection from "../../sidebar/sections/filter/FilterSection";

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

const statusTypes: any = {
  OPEN: "Open",
  SETTLED: "Settled",
  EXPIRED: "Expired",
};

function InvoicesPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const [limit, setLimit] = useLocalStorage("invoicesLimit", 100);
  const [offset, setOffset] = useState(0);

  const allViews = useGetTableViewsQuery<{
    data: AllViewsResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  const invoiceViews = allViews?.data ? allViews.data["invoices"] : [InvoiceViewTemplate];
  const [view, selectView] = useView(invoiceViews, 0);

  const invoicesResponse = useGetInvoicesQuery<{
    data: InvoicesResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(
    {
      limit: limit,
      offset: offset,
      order: view.sortBy,
    },
    { skip: !allViews.isSuccess }
  );

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  // if (invoicesResponse?.data?.data) {
  //   data = invoicesResponse?.data?.data.map((invoice: any) => {
  //     const invoice_state = statusTypes[invoice.invoice_state];
  //
  //     return {
  //       ...invoice,
  //       invoice_state,
  //     };
  //   });
  // }

  // General logic for toggling the sidebar sections
  const initialSectionState: sections = {
    filter: false,
    sort: false,
    columns: false,
  };

  const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);

  const sidebarSectionHandler = (section: keyof sections) => {
    return () => {
      setActiveSidebarSections({
        ...activeSidebarSections,
        [section]: !activeSidebarSections[section],
      });
    };
  };

  const closeSidebarHandler = () => {
    return () => {
      setSidebarExpanded(false);
    };
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <Button
          buttonColor={buttonColor.green}
          text={t.header.newInvoice}
          className={"collapse-tablet"}
          icon={<InvoiceIcon />}
          onClick={() => {
            navigate(NEW_INVOICE, { state: { background: location } });
          }}
        />
        <TableControlsButton
          onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}
          icon={OptionsIcon}
          id={"tableControlsButton"}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  // const filterColumns = clone(allColumns).map((c: any) => {
  //   switch (c.key) {
  //     case "invoiceState":
  //       c.selectOptions = Object.keys(statusTypes).map((key: any) => {
  //         return {
  //           value: key,
  //           label: statusTypes[String(key)],
  //         };
  //       });
  //       break;
  //   }
  //   return c;
  // });
  //
  // const handleFilterUpdate = (updated: Clause) => {
  //   dispatch(updateInvoicesFilters({ filters: updated.toJSON() }));
  // };
  //

  //
  // const handleSortUpdate = (updated: Array<OrderBy>) => {
  //   setOrderBy(updated);
  // };
  //
  // const updateColumnsHandler = (columns: Array<any>) => {
  //   dispatch(updateColumns({ columns: columns }));
  // };

  const sidebar = (
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SectionContainer
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={AllInvoicesColumns} view={view} />
      </SectionContainer>
      <SectionContainer
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection columns={AllInvoicesColumns} view={view} defaultFilter={DefaultFilter} />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={SortableInvoiceColumns} view={view} defaultSortBy={DefaultSortValue} />
      </SectionContainer>
    </Sidebar>
  );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/invoices"}>
      Invoices
    </Link>,
  ];
  const pagination = (
    <Pagination
      limit={limit}
      offset={offset}
      total={invoicesResponse?.data?.pagination?.total || 0}
      perPageHandler={setLimit}
      offsetHandler={setOffset}
    />
  );
  return (
    <TablePageTemplate
      title={"Invoices"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={pagination}
    >
      <Table
        cellRenderer={DefaultCellRenderer}
        data={invoicesResponse?.data?.data || []}
        activeColumns={view.columns || []}
        isLoading={invoicesResponse.isLoading || invoicesResponse.isFetching || invoicesResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default InvoicesPage;
