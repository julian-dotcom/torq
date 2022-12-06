import {
  MoneyHand20Regular as TransactionIcon,
  Options20Regular as OptionsIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import { useGetPaymentsQuery } from "./paymentsApi";
import { NEW_PAYMENT } from "constants/routes";
import Button, { buttonColor } from "components/buttons/Button";
import { FilterInterface } from "features/sidebar/sections/filter/filter";
import Table from "features/table/Table";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import { useLocation } from "react-router";
import { Link, useNavigate } from "react-router-dom";
import { PaymentsResponse } from "./types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import useTranslations from "services/i18n/useTranslations";
import {
  AllPaymentsColumns,
  DefaultPaymentView,
  PaymentsFilterTemplate,
  PaymentsSortTemplate,
  SortablePaymentsColumns,
} from "./paymentDefaults";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectPaymentsView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";

function PaymentsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPaymentsView);

  const [getPagination, limit, offset] = usePagination("invoices");

  const paymentsResponse = useGetPaymentsQuery<{
    data: PaymentsResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters ? (viewResponse.view.filters.toJSON() as FilterInterface) : undefined,
    },
    { skip: !isSuccess }
  );

  // if (paymentsResponse?.data?.data) {
  //   data = paymentsResponse?.data?.data.map((payment: any) => {
  //     const failure_reason = failureReasons[payment.failure_reason];
  //     const status = statusTypes[payment.status];
  //
  //     return {
  //       ...payment,
  //       failure_reason,
  //       status,
  //     };
  //   });
  // }

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

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
          text={"New Payment"}
          className={"collapse-tablet"}
          icon={<TransactionIcon />}
          onClick={() => {
            navigate(NEW_PAYMENT, { state: { background: location } });
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
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllPaymentsColumns}
      defaultView={DefaultPaymentView}
      filterableColumns={AllPaymentsColumns}
      filterTemplate={PaymentsFilterTemplate}
      sortableColumns={SortablePaymentsColumns}
      sortByTemplate={PaymentsSortTemplate}
    />
  );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/payments"}>
      Payments
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={"Payments"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(paymentsResponse?.data?.pagination?.total || 0)}
    >
      <Table
        cellRenderer={DefaultCellRenderer}
        data={paymentsResponse?.data?.data || []}
        activeColumns={viewResponse.view.columns || []}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
