import Table from "features/table/Table";
import { useGetInvoicesQuery } from "./invoiceApi";
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
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import { SectionContainer } from "features/section/SectionContainer";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_INVOICE } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import { InvoicesResponse } from "./invoiceTypes";
import {
  AllInvoicesColumns,
  InvoiceFilterTemplate,
  InvoiceSortTemplate,
  InvoiceViewTemplate,
  SortableInvoiceColumns,
} from "./invoiceDefaults";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";

import SortSection from "../../sidebar/sections/sort/SortSection";
import { useView } from "../../viewManagement/useView";
import FilterSection from "../../sidebar/sections/filter/FilterSection";
import { FilterInterface } from "../../sidebar/sections/filter/filter";
import { usePagination } from "../../../components/table/pagination/usePagination";

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

  const [view, selectView, isViewsLoaded] = useView("invoices", AllInvoicesColumns, 0, InvoiceViewTemplate);
  const [getPagination, limit, offset] = usePagination("invoices");

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
      filter: view.filters.length ? (view.filters.toJSON() as FilterInterface) : undefined,
    },
    { skip: !isViewsLoaded }
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
        <FilterSection columns={AllInvoicesColumns} view={view} defaultFilter={InvoiceFilterTemplate} />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={SortableInvoiceColumns} view={view} defaultSortBy={InvoiceSortTemplate} />
      </SectionContainer>
    </Sidebar>
  );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/invoices"}>
      Invoices
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={"Invoices"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(invoicesResponse?.data?.pagination?.total || 0)}
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
