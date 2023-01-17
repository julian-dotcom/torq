import { Link } from "react-router-dom";
import { Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import {
  AllHtlcColumns,
  DefaultHtlcsView,
  HtlcsFilterTemplate,
  HtlcsSortByTemplate,
} from "features/htlcs/htlcsDefaults";
import useTranslations from "services/i18n/useTranslations";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { selectHtlcsView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useGetChannelsQuery, useGetHtlcsQuery } from "apiSlice";
import { Htlc } from "features/htlcs/htlcsTypes";
import htlcsCellRenderer from "features/htlcs/htlcsCells";
import Table from "features/table/Table";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import styles from "features/htlcs/htlcs_table.module.scss";

function useHtlcsTotals(data: Array<Htlc>): Htlc | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Htlc, current: Htlc) => {
    return {
      ...prev,
      alias: "Total",
      locked: true,
      incomingAmountMsat: prev.incomingAmountMsat + current.incomingAmountMsat,
      incomingChannelCapacity: prev.incomingChannelCapacity + current.incomingChannelCapacity,
      outgoingAmountMsat: prev.outgoingAmountMsat + current.outgoingAmountMsat,
      outgoingChannelCapacity: prev.outgoingChannelCapacity + current.outgoingChannelCapacity,
    };
  });
}

function useHtlcsMaximums(data: Array<Htlc>): Htlc | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Htlc, current: Htlc) => {
    return {
      ...prev,
      alias: "Max",
      incomingAmountMsat: Math.max(prev.incomingAmountMsat, current.incomingAmountMsat),
      incomingChannelCapacity: Math.max(prev.incomingChannelCapacity, current.incomingChannelCapacity),
      outgoingAmountMsat: Math.max(prev.outgoingAmountMsat, current.outgoingAmountMsat),
      outgoingChannelCapacity: Math.max(prev.outgoingChannelCapacity, current.outgoingChannelCapacity),
    };
  });
}

function HtlcsPage() {
  const { t } = useTranslations();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();

  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { viewResponse, selectedViewIndex } = useAppSelector(selectHtlcsView);
  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const htlcsResponse = useGetHtlcsQuery<{
    data: Array<Htlc>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ from: from, to: to, network: activeNetwork }, { skip: !isSuccess });
  useGetChannelsQuery({ network: activeNetwork });

  const filteredData = useFilterData(htlcsResponse.data, viewResponse.view.filters);
  const sortedData = useSortData(filteredData, viewResponse.view.sortBy);
  const totalsRowData = useHtlcsTotals(sortedData);
  const maxRowData = useHtlcsMaximums(sortedData);

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup></TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllHtlcColumns}
      defaultView={DefaultHtlcsView}
      filterableColumns={AllHtlcColumns}
      filterTemplate={HtlcsFilterTemplate}
      sortableColumns={AllHtlcColumns}
      sortByTemplate={HtlcsSortByTemplate}
      enableGroupBy={true}
    />
  );

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"/analyse/htlcs"}>
      Htlcs
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={t.htlcs}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      titleContent={
        <div className={styles.htlcsControls}>
          <TimeIntervalSelect />
          <TableControlsButton
            onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}
            icon={OptionsIcon}
            id={"tableControlsButton"}
          />
        </div>
      }
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <Table
        activeColumns={viewResponse.view.columns || []}
        data={sortedData}
        totalRow={totalsRowData ? totalsRowData : undefined}
        maxRow={maxRowData ? maxRowData : undefined}
        cellRenderer={htlcsCellRenderer}
        isLoading={htlcsResponse.isLoading || htlcsResponse.isFetching || htlcsResponse.isUninitialized}
        showTotals={true}
      />
      {/*<HtlcsDataWrapper viewResponse={viewResponse} loadingViews={!isSuccess} />*/}
    </TablePageTemplate>
  );
}

export default HtlcsPage;
