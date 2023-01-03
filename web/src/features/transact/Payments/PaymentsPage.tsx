import { MoneyHand20Regular as TransactionIcon, Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import { useGetPaymentsQuery } from "./paymentsApi";
import { NEW_PAYMENT } from "constants/routes";
import Button, { buttonColor } from "components/buttons/Button";
import Table from "features/table/Table";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import { useLocation } from "react-router";
import { Link, useNavigate } from "react-router-dom";
import { Payment } from "./types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import {
  AllPaymentsColumns,
  DefaultPaymentView,
  FailureReasonLabels,
  FilterablePaymentsColumns,
  PaymentsFilterTemplate,
  PaymentsSortTemplate,
  SortablePaymentsColumns,
  StatusTypeLabels,
} from "./paymentDefaults";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectPaymentsView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectActiveNetwork } from "features/network/networkSlice";

function useMaximums(data: Array<Payment>): Payment | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Payment, current: Payment) => {
    return {
      ...prev,
      alias: "Max",
      paymentIndex: Math.max(prev.paymentIndex, current.paymentIndex),
      value: Math.max(prev.value, current.value),
      fee: Math.max(prev.fee, current.fee),
      ppm: Math.max(prev.ppm, current.ppm),
      paymentHash: Math.max(prev.paymentHash, current.paymentHash),
      paymentPreimage: Math.max(prev.paymentPreimage, current.paymentPreimage),
      countFailedAttempts: Math.max(prev.countFailedAttempts, current.countFailedAttempts),
      countSuccessfulAttempts: Math.max(prev.countSuccessfulAttempts, current.countSuccessfulAttempts),
      secondsInFlight: Math.max(prev.secondsInFlight, current.secondsInFlight),
    };
  });
}

function PaymentsPage() {
  const navigate = useNavigate();
  const location = useLocation();

  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPaymentsView);

  const [getPagination, limit, offset] = usePagination("invoices");

  const paymentsResponse = useGetPaymentsQuery(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters ? viewResponse.view.filters : undefined,
      network: activeNetwork,
    },
    { skip: !isSuccess }
  );

  let data = paymentsResponse.data?.data || [];
  data = data.map((payment: any) => {
    const failureReason = FailureReasonLabels.get(payment.failure_reason) || "";
    const status = StatusTypeLabels.get(payment.status) || "";

    return {
      ...payment,
      failureReason: failureReason,
      status: status,
    };
  });

  const maxRow = useMaximums(data);

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={buttonColor.green}
            text={"New Payment"}
            className={"collapse-tablet"}
            icon={<TransactionIcon />}
            onClick={() => {
              navigate(NEW_PAYMENT, { state: { background: location } });
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
      allColumns={AllPaymentsColumns}
      defaultView={DefaultPaymentView}
      filterableColumns={FilterablePaymentsColumns}
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
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
        maxRow={maxRow}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
