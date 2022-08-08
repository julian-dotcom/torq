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

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

const activeColumns: ColumnMetaData[] = [
  { key: "date", heading: "Date", type: "DateCell", valueType: "date" },
  { key: "status", heading: "Status", type: "TextCell", valueType: "string" },
  { key: "value", heading: "Value", type: "NumericCell", valueType: "number" },
  { key: "fee", heading: "Fee", type: "NumericCell", valueType: "number" },
  { key: "is_rebalance", heading: "Rebalance", type: "BooleanCell", valueType: "string" },
  { key: "seconds_in_flight", heading: "Seconds In Flight", type: "BarCell", valueType: "number" },
  { key: "failueReason", heading: "Failure Reason", type: "TextCell", valueType: "string" },
  { key: "is_mpp", heading: "MPP", type: "BooleanCell", valueType: "string" },
  // { key: "payment_hash", heading: "Payment Hash", type: "TextCell", valueType: "string" },
  // { key: "payment_index", heading: "Payment Index", type: "TextCell", valueType: "string" },
  // { key: "payment_preimage", heading: "Payment Preimage", type: "TextCell", valueType: "string" },
  // { key: "payment_request", heading: "Payment Request", type: "TextCell", valueType: "string" },
  { key: "count_failed_attempts", heading: "Failed Attempts", type: "NumericCell", valueType: "number" },
  { key: "count_successful_attempts", heading: "Successful Attempts", type: "NumericCell", valueType: "number" },
  { key: "destination_pub_key", heading: "Destination", type: "TextCell", valueType: "string" },
];

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
  const paymentsResponse = useGetPaymentsQuery({});

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const data = paymentsResponse?.data?.data.map((payment: any) => {
    const value = (payment?.value_msat || 0) / 1000;
    const fee = (payment?.fee_msat || 0) / 1000;
    const failureReason = failureReasons[payment.failure_reason];
    const status = statusTypes[payment.status];
    return { ...payment, fee, value, failureReason, status };
  });

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

  const sidebar = (
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SidebarSection
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        sectionToggleHandler={sidebarSectionHandler("columns")}
      >
        {"Something"}
      </SidebarSection>
      <SidebarSection
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        sectionToggleHandler={sidebarSectionHandler("filter")}
      >
        {"Something"}
      </SidebarSection>
      <SidebarSection
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        sectionToggleHandler={sidebarSectionHandler("sort")}
      >
        {"Something"}
      </SidebarSection>
    </Sidebar>
  );

  const breadcrumbs = ["Transactions", <Link to={"/transactions/payments"}>Payments</Link>];
  return (
    <TablePageTemplate
      title={"Payments"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <Table
        rowRenderer={rowRenderer}
        data={data || []}
        activeColumns={columns || []}
        isLoading={paymentsResponse.isLoading || paymentsResponse.isFetching || paymentsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PaymentsPage;
