import Table, { ColumnMetaData } from "features/table/Table";
import cellStyles from "features/table/cells/cell.module.scss";
import { useGetPaymentsQuery } from "apiSlice";
import { Link } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
} from "@fluentui/react-icons";
import Sidebar, { SidebarSection } from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import TransactTabs from "./TransactTabs";
import AliasCell from "../table/cells/AliasCell";
import classNames from "classnames";
import NumericCell from "../table/cells/NumericCell";
import DateCell from "../table/cells/DateCell";
import BooleanCell from "../table/cells/BooleanCell";
import BarCell from "../table/cells/BarCell";
import TextCell from "../table/cells/TextCell";
import EnumCell from "../table/cells/EnumCell";
import Pagination from "features/table/pagination/Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";
import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
import FilterSection from "../sidebar/sections/filter/FilterSection";
import { Clause, deserialiseQuery, FilterClause, FilterInterface } from "../sidebar/sections/filter/filter";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import {
  selectActiveColumns,
  selectAllColumns,
  selectPaymentsFilters,
  updateColumns,
  updatePaymentsFilters,
} from "./transactSlice";
import { FilterCategoryType } from "../sidebar/sections/filter/filter";
import ColumnsSection from "../sidebar/sections/columns/ColumnsSection";
import clone from "../../clone";
import { formatDuration, intervalToDuration } from "date-fns";
import { format } from "d3";

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

const subSecFormat = format("0.2f");

function rowRenderer(row: any, index: number, column: ColumnMetaData, columnIndex: number) {
  const key = column.key;
  switch (column.type) {
    case "AliasCell":
      return (
        <AliasCell
          current={row[key] as string}
          chanId={row["chan_id"]}
          open={row["open"]}
          className={classNames(key, index, cellStyles.locked)}
          key={key + index + columnIndex}
        />
      );
    case "NumericCell":
      return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
    case "EnumCell":
      return <EnumCell value={row[key] as string} icon={ColumnsIcon} className={key} key={key + index + columnIndex} />;
    case "DateCell":
      return <DateCell value={row[key] as string} className={key} key={key + index + columnIndex} />;
    case "BooleanCell":
      return (
        <BooleanCell
          falseTitle={"Failure"}
          trueTitle={"Success"}
          value={row[key] as boolean}
          className={classNames(key)}
          key={key + index + columnIndex}
        />
      );
    case "BarCell":
      return (
        <BarCell
          current={row[key] as number}
          previous={row[key] as number}
          total={column.max as number}
          className={key}
          key={key + index + columnIndex}
        />
      );
    case "TextCell":
      return (
        <TextCell current={row[key] as string} className={classNames(column.key, index)} key={column.key + index} />
      );
    default:
      return <NumericCell current={row[key] as number} className={key} key={key + index + columnIndex} />;
  }
}

function PaymentsPage() {
  const [limit, setLimit] = useLocalStorage("paymentsLimit", 100);
  const [offset, setOffset] = useState(0);
  const [orderBy, setOrderBy] = useLocalStorage("paymentsOrderBy", [
    {
      key: "date",
      direction: "asc",
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
      let pif = "Unknown";
      if (payment.seconds_in_flight >= 1) {
        const d = intervalToDuration({ start: 0, end: payment.seconds_in_flight * 1000 });
        pif = formatDuration({
          years: d.years,
          months: d.months,
          days: d.days,
          hours: d.hours,
          minutes: d.minutes,
          seconds: d.seconds,
        });
      } else if (payment.seconds_in_flight < 1 && payment.seconds_in_flight > 0) {
        pif = `${subSecFormat(payment.seconds_in_flight)} seconds`;
      }
      return {
        ...payment,
        failure_reason,
        status,
        seconds_in_flight: pif,
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

  const setSection = (section: keyof sections) => {
    return () => {
      if (activeSidebarSections[section] && sidebarExpanded) {
        setSidebarExpanded(false);
        setActiveSidebarSections(initialSectionState);
      } else {
        setSidebarExpanded(true);
        setActiveSidebarSections({
          ...initialSectionState,
          [section]: true,
        });
      }
    };
  };
  const sidebarSectionHandler = (section: keyof sections) => {
    return () => {
      setActiveSidebarSections({
        ...initialSectionState,
        [section]: !activeSidebarSections[section],
      });
    };
  };

  const closeSidebarHandler = () => {
    return () => {
      setSidebarExpanded(false);
      setActiveSidebarSections(initialSectionState);
    };
  };

  const tableControls = (
    <TableControlSection>
      <TransactTabs />

      <TableControlsButtonGroup>
        <TableControlsButton
          onClickHandler={setSection("columns")}
          icon={ColumnsIcon}
          active={activeSidebarSections.columns}
        />
        <TableControlsButton
          onClickHandler={setSection("filter")}
          icon={FilterIcon}
          active={activeSidebarSections.filter}
        />
        <TableControlsButton onClickHandler={setSection("sort")} icon={SortIcon} active={activeSidebarSections.sort} />
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
        rowRenderer={rowRenderer}
        data={data}
        activeColumns={columns || []}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
