import {
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Filter20Regular as FilterIcon,
  MoneyHand20Regular as TransactionIcon,
  Options20Regular as OptionsIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import { useView } from "features/viewManagement/useView";
import { useGetPaymentsQuery } from "./paymentsApi";
import { NEW_PAYMENT } from "constants/routes";
import Button, { buttonColor } from "components/buttons/Button";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import { FilterInterface } from "features/sidebar/sections/filter/filter";
import SortSection from "features/sidebar/sections/sort/SortSection";
import Sidebar from "features/sidebar/Sidebar";
import Table from "features/table/Table";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import { useLocation } from "react-router";
import { Link, useNavigate } from "react-router-dom";
import { SectionContainer } from "features/section/SectionContainer";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import { PaymentsResponse } from "./types";
import DefaultCellRenderer from "../../table/DefaultCellRenderer";
import useTranslations from "../../../services/i18n/useTranslations";
import {
  AllPaymentsColumns,
  DefaultPaymentView,
  FilterTemplate,
  PaymentsSortTemplate,
  SortablePaymentsColumns,
} from "./paymentDefaults";
import { usePagination } from "../../../components/table/pagination/usePagination";

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

function PaymentsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const [view, selectView, isViewsLoaded] = useView("payments", 0, DefaultPaymentView);
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
      order: view.sortBy,
      filter: view.filters.length ? (view.filters.toJSON() as FilterInterface) : undefined,
    },
    { skip: !isViewsLoaded }
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

  // const columns = activeColumns.map((column: ColumnMetaData<Payment>, _: number) => {
  //   if (column.type === "number") {
  //     return {
  //       ...column,
  //       max: Math.max(column.max ?? 0, data[column.key].max ?? 0),
  //     };
  //   } else {
  //     return column;
  //   }
  // });

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
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
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SectionContainer
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={AllPaymentsColumns} view={view} />
      </SectionContainer>
      <SectionContainer
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection columns={AllPaymentsColumns} view={view} defaultFilter={FilterTemplate} />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={SortablePaymentsColumns} view={view} defaultSortBy={PaymentsSortTemplate} />
      </SectionContainer>
    </Sidebar>
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
        activeColumns={view.columns}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
