import Table, { ColumnMetaData } from "features/table/Table";
import { useGetPaymentsQuery } from "apiSlice";
import { Link } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Options20Regular as OptionsIcon,
} from "@fluentui/react-icons";
import Sidebar, { SidebarSection } from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import TransactTabs from "../TransactTabs";
import Pagination from "features/table/pagination/Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";
import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
import FilterSection from "../../sidebar/sections/filter/FilterSection";
import { Clause, deserialiseQuery, FilterClause, FilterInterface } from "../../sidebar/sections/filter/filter";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import {
  selectActiveColumns,
  selectAllColumns,
  selectPaymentsFilters,
  updateColumns,
  updatePaymentsFilters,
} from "./paymentsSlice";
import { FilterCategoryType } from "features/sidebar/sections/filter/filter";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import clone from "clone";

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
  const [limit, setLimit] = useLocalStorage("paymentsLimit", 100);
  const [offset, setOffset] = useState(0);
  const [orderBy, setOrderBy] = useLocalStorage("paymentsOrderBy", [
    {
      key: "date",
      direction: "desc",
    },
  ] as Array<OrderBy>);

  const activeColumns = useAppSelector(selectActiveColumns) || [];
  const allColumns = useAppSelector(selectAllColumns);

  const dispatch = useAppDispatch();
  const filters = useAppSelector(selectPaymentsFilters);

  const paymentsResponse = useGetPaymentsQuery({
    limit: limit,
    offset: offset,
    order: orderBy,
    filter: filters && deserialiseQuery(filters).length >= 1 ? filters : undefined,
  });

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  let data: any = [];

  if (paymentsResponse?.data?.data) {
    data = paymentsResponse?.data?.data.map((payment: any) => {
      const failure_reason = failureReasons[payment.failure_reason];
      const status = statusTypes[payment.status];

      return {
        ...payment,
        failure_reason,
        status,
      };
    });
  }

  const columns = activeColumns.map((column: ColumnMetaData, index: number) => {
    if (column.type === "number") {
      return {
        ...column,
        max: Math.max(column.max ?? 0, data[column.key].max ?? 0),
      };
    } else {
      return column;
    }
  });

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
      <TransactTabs />
      <TableControlsButtonGroup>
        <TableControlsButton onClickHandler={() => setSidebarExpanded(!sidebarExpanded)} icon={OptionsIcon} />
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
      case "failure_reason":
        c.selectOptions = Object.keys(failureReasons)
          .filter((key) => key !== "FAILURE_REASON_NONE")
          .map((key: any) => {
            return {
              value: key,
              label: failureReasons[String(key)],
            };
          });
        break;
      case "status":
        c.selectOptions = Object.keys(statusTypes).map((key: any) => {
          return {
            value: key,
            label: statusTypes[String(key)],
          };
        });
    }
    return c;
  });

  const handleFilterUpdate = (updated: Clause) => {
    dispatch(updatePaymentsFilters({ filters: updated.toJSON() }));
  };

  const sortableColumns = allColumns.filter((column: ColumnMetaData) =>
    [
      "date",
      "value",
      "fee",
      "ppm",
      "status",
      "is_rebalance",
      "seconds_in_flight",
      "failure_reason",
      "is_mpp",
      "count_failed_attempts",
      "count_successful_attempts",
    ].includes(column.key)
  );

  const handleSortUpdate = (updated: Array<OrderBy>) => {
    setOrderBy(updated);
    // dispatch(updateSortBy({ sortBy: updated }));
  };

  const updateColumnsHandler = (columns: Array<any>) => {
    dispatch(updateColumns({ columns: columns }));
  };

  const sidebar = (
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SidebarSection
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={allColumns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
      </SidebarSection>
      <SidebarSection
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
      </SidebarSection>
      <SidebarSection
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={sortableColumns} orderBy={orderBy} updateHandler={handleSortUpdate} />
      </SidebarSection>
    </Sidebar>
  );

  const breadcrumbs = ["Transactions", <Link to={"/transactions/payments"}>Payments</Link>];

  const pagination = (
    <Pagination
      limit={limit}
      offset={offset}
      total={paymentsResponse?.data?.pagination?.total}
      perPageHandler={setLimit}
      offsetHandler={setOffset}
    />
  );
  return (
    <TablePageTemplate
      title={"Payments"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={pagination}
    >
      <Table
        data={data}
        activeColumns={columns || []}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
