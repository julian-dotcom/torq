import Table, { ColumnMetaData } from "features/table/Table";
import { useCreateTableViewMutation, useGetTableViewsQuery, useUpdateTableViewMutation, useGetInvoicesQuery } from "apiSlice";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Options20Regular as OptionsIcon,
  Check20Regular as InvoiceIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useEffect, useState } from "react";
import TransactTabs from "../TransactTabs";
import Pagination from "features/table/pagination/Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";
import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import { Clause, FilterInterface } from "features/sidebar/sections/filter/filter";
import { useAppDispatch, useAppSelector } from "store/hooks";
import {
  selectViews,
  updateViews,
  updateSelectedView,
  updateViewsOrder,
  DefaultView,
  selectActiveColumns,
  selectAllColumns,
  selectInvoicesFilters,
  updateColumns,
  updateInvoicesFilters,
  selectCurrentView,
  selectedViewIndex,
} from "features/transact/Invoices/invoicesSlice";
import { FilterCategoryType } from "features/sidebar/sections/filter/filter";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import clone from "clone";
import { SectionContainer } from "features/section/SectionContainer";
import Button, { buttonColor } from "features/buttons/Button";
import { NEW_INVOICE } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import { ViewResponse } from "features/viewManagement/ViewsPopover";
import { ViewInterface } from "features/table/Table";

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

  const { data: invoicesViews, isLoading } = useGetTableViewsQuery({page: 'invoices'});

  useEffect(() => {
    const views: ViewInterface[] = [];
    if (!isLoading) {
      if (invoicesViews) {
        invoicesViews?.map((v: ViewResponse) => {
          views.push(v.view)
        });

        dispatch(updateViews({ views, index: 0 }));
      } else {
        dispatch(updateViews({ views: [{...DefaultView, title: "Default View"}], index: 0 }));
      }
    }
  }, [invoicesViews, isLoading]);

  const { t } = useTranslations();
  const [limit, setLimit] = useLocalStorage("invoicesLimit", 100);
  const [offset, setOffset] = useState(0);
  const [orderBy, setOrderBy] = useLocalStorage("invoicesOrderBy", [
    {
      key: "creationDate",
      direction: "desc",
    },
  ] as OrderBy[]);

  const navigate = useNavigate();
  const location = useLocation();
  const activeColumns = useAppSelector(selectActiveColumns) || [];
  const allColumns = useAppSelector(selectAllColumns);

  const dispatch = useAppDispatch();
  const filters = useAppSelector(selectInvoicesFilters);

  const invoicesResponse = useGetInvoicesQuery({
    limit: limit,
    offset: offset,
    order: orderBy,
  });

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  let data: any = [];

  if (invoicesResponse?.data?.data) {
    data = invoicesResponse?.data?.data.map((invoice: any) => {
      const invoice_state = statusTypes[invoice.invoice_state];

      return {
        ...invoice,
        invoice_state,
      };
    });
  }

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

  const [updateTableView] = useUpdateTableViewMutation();
  const [createTableView] = useCreateTableViewMutation();
  const currentViewIndex = useAppSelector(selectedViewIndex);
  const currentView = useAppSelector(selectCurrentView);
  const saveView = () => {
    const viewMod = { ...currentView };
    viewMod.saved = true;
    if (currentView.id === undefined || null) {
      createTableView({ view: viewMod, index: currentViewIndex, page: 'invoices' });
      return;
    }
    updateTableView(viewMod);
  };

  const tableControls = (
    <TableControlSection>
      <TransactTabs
        page="invoices"
        selectViews={selectViews}
        updateViews={updateViews}
        updateSelectedView={updateSelectedView}
        selectedViewIndex={selectedViewIndex}
        updateViewsOrder={updateViewsOrder}
        DefaultView={DefaultView}
      />
      {!currentView.saved && (
        <Button
          buttonColor={buttonColor.green}
          icon={<SaveIcon />}
          text={"Save"}
          onClick={saveView}
          className={"collapse-tablet"}
        />
      )}
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

  const defaultFilter: FilterInterface = {
    funcName: "gte",
    category: "number" as FilterCategoryType,
    parameter: 0,
    key: "value",
  };

  const filterColumns = clone(allColumns).map((c: any) => {
    switch (c.key) {
      case "invoiceState":
        c.selectOptions = Object.keys(statusTypes).map((key: any) => {
          return {
            value: key,
            label: statusTypes[String(key)],
          };
        });
        break;
    }
    return c;
  });

  const handleFilterUpdate = (updated: Clause) => {
    dispatch(updateInvoicesFilters({ filters: updated.toJSON() }));
  };

  const sortableColumns = allColumns.filter((column: ColumnMetaData) =>
    [
      "creationDate",
      "settleDate",
      "invoiceState",
      "amtPaid",
      "memo",
      "value",
      "isRebalance",
      "isKeysend",
      "destinationPubKey",
      "isAmp",
      "fallbackAddr",
      "paymentAddr",
      "paymentRequest",
      "private",
      "expiry",
      "cltvExpiry",
      "updatedOn",
    ].includes(column.key)
  );

  const handleSortUpdate = (updated: Array<OrderBy>) => {
    setOrderBy(updated);
  };

  const updateColumnsHandler = (columns: Array<any>) => {
    dispatch(updateColumns({ columns: columns }));
  };

  const sidebar = (
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SectionContainer
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={allColumns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
      </SectionContainer>
      <SectionContainer
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection
          columnsMeta={filterColumns}
          filters={filters}
          filterUpdateHandler={handleFilterUpdate}
          defaultFilter={defaultFilter}
        />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={sortableColumns} orderBy={orderBy} updateHandler={handleSortUpdate} />
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
        data={data}
        activeColumns={activeColumns || []}
        isLoading={invoicesResponse.isLoading || invoicesResponse.isFetching || invoicesResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default InvoicesPage;
