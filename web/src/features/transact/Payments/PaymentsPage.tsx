import {
  MoneyHand20Regular as TransactionIcon,
  Options20Regular as OptionsIcon,
  ArrowSync20Regular as RefreshIcon,
} from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import { useGetPaymentsQuery } from "./paymentsApi";
import { NEW_PAYMENT } from "constants/routes";
import Button, { ColorVariant } from "components/buttons/Button";
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
  DefaultPaymentView,
  FailureReasonLabels,
  FilterablePaymentsColumns,
  PaymentsFilterTemplate,
  PaymentsSortTemplate,
  SortablePaymentsColumns,
  StatusTypeLabels,
} from "features/transact/Payments/paymentDefaults";
import { AllPaymentsColumns } from "features/transact/Payments/paymentsColumns.generated";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectPaymentsView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectActiveNetwork } from "features/network/networkSlice";
import useTranslations from "services/i18n/useTranslations";
import { TableResponses, ViewResponse } from "features/viewManagement/types";

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
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPaymentsView);
  const paymentsViews = useAppSelector(selectViews)("payments");
  const [updateTableView] = useUpdateTableViewMutation();

  const [getPagination, limit, offset] = usePagination("invoices");

  const paymentsResponse = useGetPaymentsQuery(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters ? viewResponse.view.filters : undefined,
      network: activeNetwork,
    },
    { skip: !isSuccess, pollingInterval: 10000 }
  );

  let data = paymentsResponse.data?.data || [];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
    mixpanel.track("Toggle Table Sidebar", { page: "Payments" });
  };

  function handleNameChange(name: string) {
    const view = paymentsViews.views[selectedViewIndex] as ViewResponse<TableResponses>;
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
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<TransactionIcon />}
            onClick={() => {
              mixpanel.track("Navigate to New Payment");
              navigate(NEW_PAYMENT, { state: { background: location } });
            }}
          >
            {t.newPayment}
          </Button>
        </TableControlsTabsGroup>
        <TableControlsButtonGroup>
          <Button
            buttonColor={ColorVariant.primary}
            icon={<RefreshIcon />}
            onClick={() => {
              mixpanel.track("Refresh Table", { page: "Payments" });
              paymentsResponse.refetch();
            }}
          />
          <TableControlsButton
            onClickHandler={() => {
              setSidebarExpanded(!sidebarExpanded);
              mixpanel.track("Toggle Table Sidebar", { page: "Payments" });
            }}
            icon={OptionsIcon}
            id={"tableControlsButton"}
          />
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
      title={viewResponse.view.title}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(paymentsResponse?.data?.pagination?.total || 0)}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
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
