import { Link } from "react-router-dom";
import { Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";
import * as Routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import Table from "features/table/Table";
import {
  ChannelsClosedFilterTemplate,
  ChannelsClosedSortTemplate,
  FilterableChannelsClosedColumns,
  SortableChannelsClosedColumns,
} from "features/channelsClosed/channelsClosedDefaults";
import { AllChannelClosedColumns } from "features/channelsClosed/channelsClosedColumns.generated";
import { useGetChannelsClosedQuery } from "apiSlice";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { selectClosedChannelView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { useState } from "react";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { useGroupBy } from "features/sidebar/sections/group/groupBy";
import { selectActiveNetwork } from "features/network/networkSlice";
import { TableResponses, ViewResponse } from "../viewManagement/types";
import { DefaultClosedChannelsView } from "./channelsClosedDefaults";
import channelsClosedCellRenderer from "./channelsClosedCellRenderer";

function useMaximums(data: Array<ChannelClosed>): ChannelClosed | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: ChannelClosed, current: ChannelClosed) => {
    return {
      ...prev,
      alias: "Max",
      capacity: Math.max(prev.capacity, current.capacity),
    };
  });
}

function ClosedChannelsPage() {
  const { t } = useTranslations();
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectClosedChannelView);
  const channelViews = useAppSelector(selectViews)("channelsClosed");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [updateTableView] = useUpdateTableViewMutation();

  const channelsResponse = useGetChannelsClosedQuery<{
    data: Array<ChannelClosed>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !isSuccess });

  const filteredData = useFilterData(channelsResponse.data, viewResponse.view.filters);
  const sortedData = useSortData(filteredData, viewResponse.view.sortBy);
  const data = useGroupBy<ChannelClosed>(sortedData, viewResponse.view.groupBy);
  const maxRow = useMaximums(data);

  // Logic for toggling the sidebar
  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    mixpanel.track("Toggle Table Sidebar", { page: "ChannelsClosed" });
  };

  function handleNameChange(name: string) {
    const view = channelViews.views[selectedViewIndex] as ViewResponse<TableResponses>;
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
        <TableControlsTabsGroup></TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <TableControlsButton
          onClickHandler={() => {
            mixpanel.track("Toggle Table Sidebar", { page: "ChannelsClosed" });
            setSidebarExpanded(!sidebarExpanded);
          }}
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
      allColumns={AllChannelClosedColumns}
      defaultView={DefaultClosedChannelsView}
      filterableColumns={FilterableChannelsClosedColumns}
      filterTemplate={ChannelsClosedFilterTemplate}
      sortableColumns={SortableChannelsClosedColumns}
      sortByTemplate={ChannelsClosedSortTemplate}
    />
  );

  const breadcrumbs = [
    <span key="b1">{t.channels}</span>,
    <Link key="b2" to={`/${Routes.CHANNELS}/${Routes.CLOSED_CHANNELS}`}>
      {t.closedChannels}
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      titleContent={""}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      onNameChange={handleNameChange}
    >
      <Table
        cellRenderer={channelsClosedCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={channelsResponse.isLoading || channelsResponse.isFetching || channelsResponse.isUninitialized}
        maxRow={maxRow}
      />
    </TablePageTemplate>
  );
}

export default ClosedChannelsPage;
