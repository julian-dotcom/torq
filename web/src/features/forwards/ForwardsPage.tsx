import { Link } from "react-router-dom";
import { Options20Regular as OptionsIcon, ArrowDownload20Regular as DownloadCsvIcon } from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import {
  DefaultForwardsView,
  ForwardsFilterTemplate,
  ForwardsSortByTemplate,
} from "features/forwards/forwardsDefaults";
import { AllForwardsColumns } from "features/forwards/forwardsColumns.generated";
import useTranslations from "services/i18n/useTranslations";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { selectForwardsView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useGetChannelsQuery, useGetForwardsQuery } from "apiSlice";
import { Forward } from "./forwardsTypes";
import forwardsCellRenderer from "./forwardsCells";
import Table from "features/table/Table";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import styles from "./forwards_table.module.scss";
import { useGroupBy } from "features/sidebar/sections/group/groupBy";
import { createCsvFile } from "utils/JsonTableToCsv";
import Button, { ColorVariant } from "components/buttons/Button";
import { TableResponses, ViewResponse } from "features/viewManagement/types";
import { userEvents } from "utils/userEvents";

function useForwardsTotals(data: Array<Forward>): Forward | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Forward, current: Forward) => {
    return {
      ...prev,
      alias: "Total",
      locked: true,
      capacity: prev.capacity + current.capacity,
      amountIn: prev.amountIn + current.amountIn,
      amountOut: prev.amountOut + current.amountOut,
      amountTotal: prev.amountTotal + current.amountTotal,
      revenueOut: prev.revenueOut + current.revenueOut,
      revenueIn: prev.revenueIn + current.revenueIn,
      revenueTotal: prev.revenueTotal + current.revenueTotal,
      countOut: prev.countOut + current.countOut,
      countIn: prev.countIn + current.countIn,
      countTotal: prev.countTotal + current.countTotal,
      turnoverOut: prev.turnoverOut + current.turnoverOut,
      turnoverIn: prev.turnoverIn + current.turnoverIn,
      turnoverTotal: prev.turnoverTotal + current.turnoverTotal,
    };
  });
}

function useForwardsMaximums(data: Array<Forward>): Forward | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: Forward, current: Forward) => {
    return {
      ...prev,
      alias: "Max",
      capacity: Math.max(prev.capacity, current.capacity),
      amountIn: Math.max(prev.amountIn, current.amountIn),
      amountOut: Math.max(prev.amountOut, current.amountOut),
      amountTotal: Math.max(prev.amountTotal, current.amountTotal),
      revenueOut: Math.max(prev.revenueOut, current.revenueOut),
      revenueIn: Math.max(prev.revenueIn, current.revenueIn),
      revenueTotal: Math.max(prev.revenueTotal, current.revenueTotal),
      countOut: Math.max(prev.countOut, current.countOut),
      countIn: Math.max(prev.countIn, current.countIn),
      countTotal: Math.max(prev.countTotal, current.countTotal),
      turnoverOut: Math.max(prev.turnoverOut, current.turnoverOut),
      turnoverIn: Math.max(prev.turnoverIn, current.turnoverIn),
      turnoverTotal: Math.max(prev.turnoverTotal, current.turnoverTotal),
    };
  });
}

function ForwardsPage() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const [updateTableView] = useUpdateTableViewMutation();

  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { viewResponse, selectedViewIndex } = useAppSelector(selectForwardsView);
  const forwardsViews = useAppSelector(selectViews)("forwards");
  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const forwardsResponse = useGetForwardsQuery<{
    data: Array<Forward>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ from: from, to: to, network: activeNetwork }, { skip: !isSuccess });
  useGetChannelsQuery({ network: activeNetwork });

  function handleNameChange(name: string) {
    const view = forwardsViews.views[selectedViewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: name },
      });
    }
  }

  const filteredData = useFilterData(forwardsResponse.data, viewResponse.view.filters);
  const sortedData = useSortData(filteredData, viewResponse.view.sortBy);
  const data = useGroupBy<Forward>(sortedData, viewResponse.view.groupBy);
  const totalsRowData = useForwardsTotals(data);
  const maxRowData = useForwardsMaximums(data);

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    track("Toggle Table Sidebar", { page: "Forwards" });
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
      allColumns={AllForwardsColumns}
      defaultView={DefaultForwardsView}
      filterableColumns={AllForwardsColumns}
      filterTemplate={ForwardsFilterTemplate}
      sortableColumns={AllForwardsColumns}
      sortByTemplate={ForwardsSortByTemplate}
      enableGroupBy={true}
    />
  );

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"/analyse/forwards"}>
      Forwards
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
      titleContent={
        <div className={styles.forwardsControls}>
          <TimeIntervalSelect />
          <Button
            buttonColor={ColorVariant.primary}
            title={t.download}
            hideMobileText={true}
            icon={<DownloadCsvIcon />}
            onClick={() => {
              track("Downloads Table as CSV", {
                downloadTablePage: "Forwards",
                downloadTableViewTitle: viewResponse.view.title,
                downloadTableColumns: viewResponse.view.columns,
                downloadTableFilters: viewResponse.view.filters,
                downloadTableSortBy: viewResponse.view.sortBy,
              });
              createCsvFile(data, viewResponse.view.title || "Forwards");
            }}
          />
          <TableControlsButton
            onClickHandler={() => {
              setSidebarExpanded(!sidebarExpanded);
              track("Toggle Table Sidebar", { page: "Forwards" });
            }}
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
        data={data}
        totalRow={totalsRowData ? totalsRowData : undefined}
        maxRow={maxRowData ? maxRowData : undefined}
        cellRenderer={forwardsCellRenderer}
        isLoading={forwardsResponse.isLoading || forwardsResponse.isFetching || forwardsResponse.isUninitialized}
        showTotals={true}
      />
      {/*<ForwardsDataWrapper viewResponse={viewResponse} loadingViews={!isSuccess} />*/}
    </TablePageTemplate>
  );
}

export default ForwardsPage;
