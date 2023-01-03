import Table from "features/table/Table";
import { useGetInvoicesQuery } from "./invoiceApi";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  Options20Regular as OptionsIcon,
  Check20Regular as InvoiceIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_INVOICE } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import { Invoice } from "./invoiceTypes";
import {
  AllInvoicesColumns,
  InvoiceFilterTemplate,
  InvoiceSortTemplate,
  DefaultInvoiceView,
  SortableInvoiceColumns,
  FilterableInvoiceColumns,
} from "./invoiceDefaults";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectInvoicesView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectActiveNetwork } from "features/network/networkSlice";

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

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectInvoicesView);

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
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={buttonColor.green}
            text={t.header.newInvoice}
            className={"collapse-tablet"}
            icon={<InvoiceIcon />}
            onClick={() => {
              navigate(NEW_INVOICE, { state: { background: location } });
            }}
          />
        </TableControlsTabsGroup>
        <TableControlsButton
          onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}
          icon={OptionsIcon}
          id={"tableControlsButton"}
        />
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
      title={"Invoices"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(invoicesResponse?.data?.pagination?.total || 0)}
    >
      <Table
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
