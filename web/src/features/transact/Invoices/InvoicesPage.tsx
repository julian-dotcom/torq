import Table from "features/table/Table";
import { useGetInvoicesQuery } from "./invoiceApi";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  Options20Regular as OptionsIcon,
  Add20Regular as InvoiceIcon,
  ArrowSync20Regular as RefreshIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Button, { ColorVariant } from "components/buttons/Button";
import { NEW_INVOICE } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import { Invoice } from "features/transact/Invoices/invoiceTypes";
import {
  InvoiceFilterTemplate,
  InvoiceSortTemplate,
  DefaultInvoiceView,
  SortableInvoiceColumns,
  FilterableInvoiceColumns,
} from "features/transact/Invoices/invoiceDefaults";
import { AllInvoicesColumns } from "features/transact/Invoices/invoicesColumns.generated";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectInvoicesView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectActiveNetwork } from "features/network/networkSlice";
import { TableResponses, ViewResponse } from "features/viewManagement/types";
import { userEvents } from "utils/userEvents";

function useMaximums(data: Array<Invoice>): Invoice | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Invoice, current: Invoice) => {
    return {
      ...prev,
      alias: "Max",
      addIndex: Math.max(prev.addIndex, current.addIndex),
      settleIndex: Math.max(prev.settleIndex, current.settleIndex),
      value: Math.max(prev.value, current.value),
      amtPaid: Math.max(prev.amtPaid, current.amtPaid),
      expiry: Math.max(prev.expiry, current.expiry),
      cltvExpiry: Math.max(prev.cltvExpiry, current.cltvExpiry),
    };
  });
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const statusTypes: any = {
  OPEN: "Open",
  SETTLED: "Settled",
  EXPIRED: "Expired",
  CANCELED: "Canceled",
};

function InvoicesPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();
  const { track } = userEvents();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectInvoicesView);
  const invoiceView = useAppSelector(selectViews)("invoices");
  const [updateTableView] = useUpdateTableViewMutation();

  const [getPagination, limit, offset] = usePagination("invoices");
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const invoicesResponse = useGetInvoicesQuery(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters ? viewResponse.view.filters : undefined,
      network: activeNetwork,
    },
    { skip: !isSuccess }
  );

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  let data: Array<Invoice> = [];
  if (invoicesResponse?.data?.data) {
    data = invoicesResponse?.data?.data.map((invoice: Invoice) => {
      const invoiceState = statusTypes[invoice.invoiceState];

      return {
        ...invoice,
        invoiceState,
      };
    });
  }

  const maxrow = useMaximums(data || []);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    track("Toggle Table Sidebar", { page: "Invoices" });
  };

  function handleNameChange(name: string) {
    const view = invoiceView.views[selectedViewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: name },
      });
    }
  }

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            data-interom-target="new-invoice"
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<InvoiceIcon />}
            onClick={() => {
              navigate(NEW_INVOICE, { state: { background: location } });
              track("Navigate to New Invoice");
            }}
          >
            {t.header.newInvoice}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <TableControlsButtonGroup>
          <Button
            data-interom-target="refresh-table"
            buttonColor={ColorVariant.primary}
            icon={<RefreshIcon />}
            onClick={() => {
              track("Refresh Table", { page: "Invoices" });
              invoicesResponse.refetch();
            }}
          />
          <Button
            data-interom-target="table-settings"
            onClick={() => {
              setSidebarExpanded(!sidebarExpanded);
              track("Toggle Table Sidebar", {
                page: "Invoices",
              });
            }}
            icon={<OptionsIcon />}
            hideMobileText={true}
            id={"tableControlsButton"}
          >
            {t.Options}
          </Button>
        </TableControlsButtonGroup>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllInvoicesColumns}
      defaultView={DefaultInvoiceView}
      filterableColumns={FilterableInvoiceColumns}
      filterTemplate={InvoiceFilterTemplate}
      sortableColumns={SortableInvoiceColumns}
      sortByTemplate={InvoiceSortTemplate}
    />
  );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/invoices"}>
      Invoices
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(invoicesResponse?.data?.pagination?.total || 0)}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
    >
      <Table
        intercomTarget={"invoices-table"}
        cellRenderer={DefaultCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={invoicesResponse.isLoading || invoicesResponse.isFetching || invoicesResponse.isUninitialized}
        maxRow={maxrow}
      />
    </TablePageTemplate>
  );
}

export default InvoicesPage;
