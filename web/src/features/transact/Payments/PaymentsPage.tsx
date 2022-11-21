import {
  // ArrowSortDownLines20Regular as SortIcon,
  // ColumnTriple20Regular as ColumnsIcon,
  // Filter20Regular as FilterIcon,
  MoneyHand20Regular as TransactionIcon,
  // Options20Regular as OptionsIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import {
  useGetTableViewsQuery,
  // useCreateTableViewMutation,
  // useUpdateTableViewMutation,
} from "features/viewManagement/viewsApiSlice";
import { useGetPaymentsQuery } from "apiSlice";
// import clone from "clone";
import { NEW_PAYMENT } from "constants/routes";
import Button, { buttonColor } from "components/buttons/Button";
import useLocalStorage from "features/helpers/useLocalStorage";
// import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
// import { FilterCategoryType } from "features/sidebar/sections/filter/filter";
// import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
// import Sidebar from "features/sidebar/Sidebar";
import Pagination from "components/table/pagination/Pagination";
import Table from "features/table/Table";
// import { ColumnMetaData } from "features/table/types";
import TablePageTemplate, {
  TableControlsButtonGroup,
  TableControlSection,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import { useLocation } from "react-router";
import { Link, useNavigate } from "react-router-dom";
// import { useAppDispatch, useAppSelector } from "store/hooks";
// import { SectionContainer } from "features/section/SectionContainer";
// import { Clause, FilterInterface } from "features/sidebar/sections/filter/filter";
// import FilterSection from "features/sidebar/sections/filter/FilterSection";
import { AllViewsResponse } from "features/viewManagement/types";
import { PaymentsResponse } from "./types";
import DefaultCellRenderer from "../../table/DefaultCellRenderer";
import useTranslations from "../../../services/i18n/useTranslations";
import { DefaultPaymentView } from "./paymentDefaults";

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

const statusTypes: any = {
  SUCCEEDED: "Succeeded",
  FAILED: "Failed",
  IN_FLIGHT: "In Flight",
};

const failureReasons: any = {
  FAILURE_REASON_NONE: "",
  FAILURE_REASON_TIMEOUT: "Timeout",
  FAILURE_REASON_NO_ROUTE: "No Route",
  FAILURE_REASON_ERROR: "Error",
  FAILURE_REASON_INCORRECT_PAYMENT_DETAILS: "Incorrect Payment Details",
  FAILURE_REASON_INCORRECT_PAYMENT_AMOUNT: "Incorrect Payment Amount",
  FAILURE_REASON_PAYMENT_HASH_MISMATCH: "Payment Hash Mismatch",
  FAILURE_REASON_INCORRECT_PAYMENT_REQUEST: "Incorrect Payment Request",
  FAILURE_REASON_UNKNOWN: "Unknown",
};

function PaymentsPage() {
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
  const views = allViews?.data ? allViews.data["payments"] : [DefaultPaymentView];
  const [selectedView, setSelectedView] = useState(0);

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
      // order: invoiceViews[selectedView].sortBy,
    },
    { skip: !allViews.isSuccess }
  );

  // useEffect(() => {
  //   const views: ViewInterface<Payment>[] = [];
  //   if (!isLoading) {
  //     if (paymentsViews) {
  //       paymentsViews?.map((v: ViewResponse<Payment>) => {
  //         views.push(v.view);
  //       });
  //
  //       dispatch(updateViews({ views, index: 0 }));
  //     } else {
  //       dispatch(updateViews({ views: [{ ...DefaultView, title: "Default View" }], index: 0 }));
  //     }
  //   }
  // }, [paymentsViews, isLoading]);
  //
  // const [limit, setLimit] = useLocalStorage("paymentsLimit", 100);
  // const [offset, setOffset] = useState(0);
  // const [orderBy, setOrderBy] = useLocalStorage("paymentsOrderBy", [
  //   {
  //     key: "date",
  //     direction: "desc",
  //   },
  // ] as OrderBy[]);
  //
  // const activeColumns = useAppSelector(selectActiveColumns) || [];
  // const allColumns = useAppSelector(selectAllColumns);
  //
  // const dispatch = useAppDispatch();
  // const navigate = useNavigate();
  // const filters = useAppSelector(selectPaymentsFilters);
  //
  // const paymentsResponse = useGetPaymentsQuery({
  //   limit: limit,
  //   offset: offset,
  //   order: orderBy,
  // });
  //
  // // Logic for toggling the sidebar
  // const [sidebarExpanded, setSidebarExpanded] = useState(false);
  // let data: any = [];
  //
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
  //
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
  //
  // // General logic for toggling the sidebar sections
  // const initialSectionState: sections = {
  //   filter: false,
  //   sort: false,
  //   columns: false,
  // };
  //
  // const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);
  //
  // const sidebarSectionHandler = (section: keyof sections) => {
  //   return () => {
  //     setActiveSidebarSections({
  //       ...activeSidebarSections,
  //       [section]: !activeSidebarSections[section],
  //     });
  //   };
  // };
  //
  // const closeSidebarHandler = () => {
  //   return () => {
  //     setSidebarExpanded(false);
  //   };
  // };
  //
  // const location = useLocation();
  //
  // const [updateTableView] = useUpdateTableViewMutation();
  // const [createTableView] = useCreateTableViewMutation();
  // const currentViewIndex = useAppSelector(selectedViewIndex);
  // const currentView = useAppSelector(selectCurrentView);
  // const saveView = () => {
  //   const viewMod = { ...currentView };
  //   viewMod.saved = true;
  //   if (currentView.id === undefined || null) {
  //     createTableView({ view: viewMod, index: currentViewIndex, page: "payments" });
  //     return;
  //   }
  //   updateTableView(viewMod);
  // };

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
        {/*<TableControlsButton*/}
        {/*  onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}*/}
        {/*  icon={OptionsIcon}*/}
        {/*  id={"tableControlsButton"}*/}
        {/*/>*/}
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  // const defaultFilter: FilterInterface = {
  //   funcName: "gte",
  //   category: "number" as FilterCategoryType,
  //   parameter: 0,
  //   key: "value",
  // };

  // const filterColumns = clone(allColumns).map((c: any) => {
  //   switch (c.key) {
  //     case "failureReason":
  //       c.selectOptions = Object.keys(failureReasons)
  //         .filter((key) => key !== "FAILURE_REASON_NONE")
  //         .map((key: any) => {
  //           return {
  //             value: key,
  //             label: failureReasons[String(key)],
  //           };
  //         });
  //       break;
  //     case "status":
  //       c.selectOptions = Object.keys(statusTypes).map((key: any) => {
  //         return {
  //           value: key,
  //           label: statusTypes[String(key)],
  //         };
  //       });
  //   }
  //   return c;
  // });

  // const handleFilterUpdate = (updated: Clause) => {
  //   dispatch(updatePaymentsFilters({ filters: updated.toJSON() }));
  // };
  //
  // const sortableColumns = allColumns.filter((column: ColumnMetaData<Payment>) =>
  //   [
  //     "date",
  //     "value",
  //     "fee",
  //     "ppm",
  //     "status",
  //     "isRebalance",
  //     "secondsInFlight",
  //     "failureReason",
  //     "isMpp",
  //     "countFailedAttempts",
  //     "countSuccessfulAttempts",
  //   ].includes(column.key)
  // );

  // const handleSortUpdate = (updated: Array<OrderBy>) => {
  //   setOrderBy(updated);
  //   // dispatch(updateSortBy({ sortBy: updated }));
  // };
  //
  // const updateColumnsHandler = (columns: ColumnMetaData<Payment>[]) => {
  //   dispatch(updateColumns({ columns: columns }));
  // };

  // const sidebar = (
  //   <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
  //     <SectionContainer
  //       title={"Columns"}
  //       icon={ColumnsIcon}
  //       expanded={activeSidebarSections.columns}
  //       handleToggle={sidebarSectionHandler("columns")}
  //     >
  //       <ColumnsSection columns={allColumns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
  //     </SectionContainer>
  //     <SectionContainer
  //       title={"Filter"}
  //       icon={FilterIcon}
  //       expanded={activeSidebarSections.filter}
  //       handleToggle={sidebarSectionHandler("filter")}
  //     >
  //       <FilterSection
  //         columnsMeta={filterColumns}
  //         filters={filters}
  //         filterUpdateHandler={handleFilterUpdate}
  //         defaultFilter={defaultFilter}
  //       />
  //     </SectionContainer>
  //     <SectionContainer
  //       title={"Sort"}
  //       icon={SortIcon}
  //       expanded={activeSidebarSections.sort}
  //       handleToggle={sidebarSectionHandler("sort")}
  //     >
  //       <SortSection columns={sortableColumns} orderBy={orderBy} updateHandler={handleSortUpdate} />
  //     </SectionContainer>
  //   </Sidebar>
  // );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/payments"}>
      Payments
    </Link>,
  ];

  const pagination = (
    <Pagination
      limit={limit}
      offset={offset}
      total={paymentsResponse?.data?.pagination?.total || 0}
      perPageHandler={setLimit}
      offsetHandler={setOffset}
    />
  );
  return (
    <TablePageTemplate
      title={"Payments"}
      breadcrumbs={breadcrumbs}
      // sidebarExpanded={sidebarExpanded}
      // sidebar={sidebar}
      tableControls={tableControls}
      pagination={pagination}
    >
      <>
        <Table
          cellRenderer={DefaultCellRenderer}
          data={paymentsResponse?.data?.data || []}
          activeColumns={views[selectedView].columns || []}
          isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
        />
      </>
    </TablePageTemplate>
  );
}

export default PaymentsPage;
